package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/account_mgr/internal/consts"
	"github.com/starslipay/account_mgr/internal/svc"
	"github.com/starslipay/account_mgr/internal/xerr"
	"github.com/starslipay/account_mgr/model/mysql"
	"github.com/starslipay/paycomm/xerror"
	"google.golang.org/grpc/codes"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type C2CTransferMessage struct {
	TransactionId string `json:"transaction_id"`
	BuyerUid      int64  `json:"buyer_uid"`
	BuyerUserId   string `json:"buyer_user_id"`
	SellerUid     int64  `json:"seller_uid"`
	SellerUserId  string `json:"seller_user_id"`
	Amount        int64  `json:"amount"`
	CurType       int32  `json:"cur_type"`
	Desc          string `json:"desc"`
}

type C2CFinalLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewC2CFinalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *C2CFinalLogic {
	return &C2CFinalLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *C2CFinalLogic) CheckInputParams(in *account_mgr_pb.C2CReq) error {
	if in.BuyerUid <= 0 {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "buyer_uid is invalid")
	}
	if in.SellerUid <= 0 {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "seller_uid is invalid")
	}
	if in.BuyerUid == in.SellerUid {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "buyer and seller cannot be the same")
	}
	if in.Amount <= 0 {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "amount must be positive")
	}
	if in.TransactionId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "transaction_id is required")
	}
	return nil
}

func (l *C2CFinalLogic) C2CFinal(in *account_mgr_pb.C2CReq) (*account_mgr_pb.C2CRsp, error) {
	err := l.CheckInputParams(in)
	if err != nil {
		return nil, err
	}

	// 检查是否是重入
	bill, err := l.svcCtx.TC2cBillModelMaster.FindOne(l.ctx, in.TransactionId)
	if err == nil {
		if bill.BuyerUid != in.BuyerUid ||
			bill.SellerUid != in.SellerUid ||
			bill.BuyerUserId != in.BuyerUserId ||
			bill.SellerUserId != in.SellerUserId ||
			bill.Amount != in.Amount {
			return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeRepeatButInfoNotConsistent, "repeat but info not consistent")
		}
		return &account_mgr_pb.C2CRsp{
			TransactionId: in.TransactionId,
			BuyerUid:      in.BuyerUid,
			BuyerUserId:   in.BuyerUserId,
			SellerUid:     in.SellerUid,
			SellerUserId:  in.SellerUserId,
			PayTime:       bill.PayTime.Format("2006-01-02 15:04:05"),
			IsRepeat:      1,
		}, nil
	}

	var result *account_mgr_pb.C2CRsp
	err = l.svcCtx.SqlMasterConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		tcAccountModel := mysql.NewTCAccountModel(sqlx.NewSqlConnFromSession(session))
		tcAccountLogModel := mysql.NewTCAccountLogModel(sqlx.NewSqlConnFromSession(session))
		tc2cBillModel := mysql.NewTC2cBillModel(sqlx.NewSqlConnFromSession(session))
		tLocalMessageModel := mysql.NewTLocalMessageModel(sqlx.NewSqlConnFromSession(session))

		buyerAccount, err := tcAccountModel.FindOne(ctx, in.BuyerUid)
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("find buyer account failed: %v", err))
		}

		if buyerAccount.Balance < in.Amount {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeBalanceNotEnough, "balance not enough")
		}

		payTime := time.Now()

		err = tcAccountModel.SubBalance(ctx, in.BuyerUid, in.Amount)
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("sub balance failed: %v", err))
		}

		_, err = tcAccountLogModel.Insert(ctx, &mysql.TCAccountLog{
			Uid:                in.BuyerUid,
			UserId:             in.BuyerUserId,
			CounterpartyUserId: in.BuyerUserId,
			CounterpartyUid:    in.BuyerUid,
			TransactionId:      in.TransactionId,
			InoutType:          consts.InoutTypeOut,
			BizType:            consts.BizTypeC2C,
			Amount:             in.Amount,
			Balance:            buyerAccount.Balance - in.Amount,
			Desc:               in.Desc,
		})
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("insert account log failed: %v", err))
		}

		_, err = tc2cBillModel.Insert(ctx, &mysql.TC2cBill{
			TransactionId: in.TransactionId,
			BuyerUid:      in.BuyerUid,
			SellerUid:     in.SellerUid,
			BuyerUserId:   in.BuyerUserId,
			SellerUserId:  in.SellerUserId,
			Amount:        in.Amount,
			State:         consts.C2CBillStateOK,
			BizType:       consts.BizTypeC2C,
			Desc:          in.Desc,
			PayTime:       payTime,
		})
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("insert c2c bill failed: %v", err))
		}

		msgBody := &C2CTransferMessage{
			TransactionId: in.TransactionId,
			BuyerUid:      in.BuyerUid,
			BuyerUserId:   in.BuyerUserId,
			SellerUid:     in.SellerUid,
			SellerUserId:  in.SellerUserId,
			Amount:        in.Amount,
			CurType:       in.CurType,
			Desc:          in.Desc,
		}
		bodyBytes, err := json.Marshal(msgBody)
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, fmt.Sprintf("marshal message failed: %v", err))
		}

		_, err = tLocalMessageModel.Insert(ctx, &mysql.TLocalMessage{
			TransactionId: in.TransactionId,
			MsgType:       consts.MsgTypeC2CTransfer,
			Topic:         l.svcCtx.Config.Kafka.Topic,
			Key:           in.TransactionId,
			Body:          string(bodyBytes),
			State:         consts.MsgStateInit,
			SendCount:     0,
			MaxSendCount:  3,
			NextSendTime:  time.Now(),
		})
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("insert local message failed: %v", err))
		}

		result = &account_mgr_pb.C2CRsp{
			TransactionId: in.TransactionId,
			BuyerUid:      in.BuyerUid,
			BuyerUserId:   in.BuyerUserId,
			SellerUid:     in.SellerUid,
			SellerUserId:  in.SellerUserId,
			PayTime:       payTime.Format("2006-01-02 15:04:05"),
			IsRepeat:      0,
		}

		return nil
	})

	if err != nil {
		l.Errorf("C2CFinal transaction failed: %v", err)
		return nil, err
	}

	// 发cmq异步入账消息
	err = l.sendKafkaMessage(in.TransactionId)
	if err != nil {
		l.Errorf("send kafka message failed: %v", err)
	}

	return result, nil
}

func (l *C2CFinalLogic) sendKafkaMessage(transactionId string) error {
	msg, err := l.svcCtx.TLocalMessageModelMaster.FindOneByTransactionId(l.ctx, transactionId)
	if err != nil {
		return err
	}

	if msg.State != consts.MsgStateInit {
		return nil
	}

	_, _, err = l.svcCtx.KafkaProducer.SendMessage(&sarama.ProducerMessage{
		Topic: msg.Topic,
		Key:   sarama.StringEncoder(msg.Key),
		Value: sarama.StringEncoder(msg.Body),
	})
	if err != nil {
		newSendCount := msg.SendCount + 1
		if newSendCount >= msg.MaxSendCount {
			msg.State = consts.MsgStateDone
		} else {
			msg.SendCount = newSendCount
			msg.NextSendTime = time.Now().Add(time.Minute * time.Duration(newSendCount*2))
		}
		l.svcCtx.TLocalMessageModelMaster.Update(l.ctx, msg)
		return err
	}

	msg.State = consts.MsgStateSent
	msg.SendCount = msg.SendCount + 1
	err = l.svcCtx.TLocalMessageModelMaster.Update(l.ctx, msg)
	return err
}
