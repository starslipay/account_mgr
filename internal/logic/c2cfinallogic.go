package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

type PendingC2CTransferMessage struct {
	TransactionId string `json:"transaction_id"`
	BuyerUid      int64  `json:"buyer_uid"`
	BuyerUserId   string `json:"buyer_user_id"`
	SellerUid     int64  `json:"seller_uid"`
	SellerUserId  string `json:"seller_user_id"`
	Amount        int64  `json:"amount"`
	BizType       int32  `json:"biz_type"`
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
	bill, _ := l.svcCtx.TC2cBillModelMaster.FindOne(l.ctx, in.TransactionId)
	if bill != nil {
		if bill.BuyerUid != in.BuyerUid ||
			bill.SellerUid != in.SellerUid ||
			bill.BuyerUserId != in.BuyerUserId ||
			bill.SellerUserId != in.SellerUserId ||
			bill.Amount != in.Amount {
			return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeRepeatButInfoNotConsistent, "repeat but info not consistent")
		}
		return &account_mgr_pb.C2CRsp{
			TransactionId: bill.TransactionId,
			BuyerUid:      bill.BuyerUid,
			BuyerUserId:   bill.BuyerUserId,
			SellerUid:     bill.SellerUid,
			SellerUserId:  bill.SellerUserId,
			PayTime:       bill.PayTime.Format("2006-01-02 15:04:05"),
			IsRepeat:      1,
		}, nil
	}

	var result *account_mgr_pb.C2CRsp
	err = l.svcCtx.SqlMasterConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		tcAccountModel := mysql.NewTCAccountModel(sqlx.NewSqlConnFromSession(session))
		tcAccountLogModel := mysql.NewTCAccountLogModel(sqlx.NewSqlConnFromSession(session))
		tc2cBillModel := mysql.NewTC2cBillModel(sqlx.NewSqlConnFromSession(session))
		tPendingC2cTransferModel := mysql.NewTPendingC2cTransferModel(sqlx.NewSqlConnFromSession(session))

		buyerAccount, err := tcAccountModel.FindOneForUpdate(ctx, in.BuyerUid)
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
			CounterpartyUserId: in.SellerUserId,
			CounterpartyUid:    in.SellerUid,
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

		_, err = tPendingC2cTransferModel.Insert(ctx, &mysql.TPendingC2cTransfer{
			TransactionId: in.TransactionId,
			BuyerUid:      in.BuyerUid,
			SellerUid:     in.SellerUid,
			BuyerUserId:   in.BuyerUserId,
			SellerUserId:  in.SellerUserId,
			Amount:        in.Amount,
			State:         consts.PendingC2cTransferInit,
			BizType:       consts.BizTypeC2C,
			Desc:          in.Desc,
		})
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("insert pending c2c transfer failed: %v", err))
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

	// 异步发送cmq消息
	go l.sendCmq(in)

	return result, nil
}

func (l *C2CFinalLogic) sendCmq(in *account_mgr_pb.C2CReq) {
	message := &PendingC2CTransferMessage{
		TransactionId: in.TransactionId,
		BuyerUid:      in.BuyerUid,
		SellerUid:     in.SellerUid,
		BuyerUserId:   in.BuyerUserId,
		SellerUserId:  in.SellerUserId,
		Amount:        in.Amount,
		BizType:       consts.BizTypeC2C,
		Desc:          in.Desc,
	}
	jsonStr, err := json.Marshal(message)
	if err != nil {
		logx.Errorf("marshal message failed: %v", err)
		return
	}
	err = l.svcCtx.C2CAsyncTransferProducer.Push(l.ctx, string(jsonStr))
	if err != nil {
		logx.Errorf("send cmq message failed: %v", err)
	}
}
