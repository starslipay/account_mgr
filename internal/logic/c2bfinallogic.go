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

const (
	ModeNormal = 0 // 正常模式
	ModeSupply = 1 // 补单模式
)

type PendingC2BTransferMessage struct {
	TransactionId string `json:"transaction_id"`
}

type C2BFinalLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewC2BFinalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *C2BFinalLogic {
	return &C2BFinalLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *C2BFinalLogic) CheckInputParams(in *account_mgr_pb.C2BReq) error {
	if in.Uid <= 0 {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "uid is invalid")
	}
	if in.MerchantUid <= 0 {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "merchant_uid is invalid")
	}
	if in.Uid == in.MerchantUid {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "user and merchant cannot be the same")
	}
	if in.Amount <= 0 {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "amount must be positive")
	}
	if in.TransactionId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "transaction_id is required")
	}
	return nil
}

func (l *C2BFinalLogic) C2BFinal(in *account_mgr_pb.C2BReq) (*account_mgr_pb.C2BRsp, error) {
	err := l.CheckInputParams(in)
	if err != nil {
		return nil, err
	}

	// 检查是否是重入
	bill, _ := l.svcCtx.TC2bBillModelMaster.FindOne(l.ctx, in.TransactionId)
	if bill != nil {
		// 检查单据状态是否是OK
		if consts.C2BBillStateSuccess == bill.State {
			if bill.Uid != in.Uid ||
				bill.UserId != in.UserId ||
				bill.MerchantUid != in.MerchantUid ||
				bill.MerchantId != in.MerchantId ||
				bill.Amount != fmt.Sprintf("%d", in.Amount) {
				return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeRepeatButInfoNotConsistent, "repeat but info not consistent")
			}

			return &account_mgr_pb.C2BRsp{
				TransactionId: bill.TransactionId,
				Uid:           bill.Uid,
				UserId:        bill.UserId,
				MerchantUid:   bill.MerchantUid,
				MerchantId:    bill.MerchantId,
				PayTime:       bill.PayTime.Format("2006-01-02 15:04:05"),
				IsRepeat:      1,
			}, nil
		} else if consts.C2BBillStateClose == bill.State {
			return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeC2BBillStateAlreadyClose, "c2b bill already close")
		} else {
			return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeC2BBillStateInvalid, "c2b bill state is invalid")
		}
	}

	// 补偿模式, 单不存在返回错误
	if ModeSupply == in.Mode {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeSupplyModeC2BBillNotFound, "c2b bill not exist")
	} else if ModeNormal == in.Mode {
		// 继续流程
	} else {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "mode is not support")
	}

	var result *account_mgr_pb.C2BRsp
	err = l.svcCtx.SqlMasterConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		tcAccountModel := mysql.NewTCAccountModel(sqlx.NewSqlConnFromSession(session))
		tcAccountLogModel := mysql.NewTCAccountLogModel(sqlx.NewSqlConnFromSession(session))
		tc2bBillModel := mysql.NewTC2bBillModel(sqlx.NewSqlConnFromSession(session))
		tc2bPendingTransferModel := mysql.NewTC2bPendingTransferModel(sqlx.NewSqlConnFromSession(session))

		userAccount, err := tcAccountModel.FindOneForUpdate(ctx, in.Uid)
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("find user account failed: %v", err))
		}

		if userAccount.Balance < in.Amount {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeBalanceNotEnough, "balance not enough")
		}

		payTime := time.Now()

		err = tcAccountModel.SubBalance(ctx, in.Uid, in.Amount)
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("sub balance failed: %v", err))
		}

		_, err = tcAccountLogModel.Insert(ctx, &mysql.TCAccountLog{
			Uid:             in.Uid,
			UserId:          in.UserId,
			CounterpartyId:  in.MerchantId,
			CounterpartyUid: in.MerchantUid,
			TransactionId:   in.TransactionId,
			InoutType:       consts.InoutTypeOut,
			BizType:         consts.BizTypeC2B,
			Amount:          in.Amount,
			Balance:         userAccount.Balance - in.Amount,
			Desc:            in.Desc,
		})
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("insert account log failed: %v", err))
		}

		_, err = tc2bBillModel.Insert(ctx, &mysql.TC2bBill{
			TransactionId: in.TransactionId,
			Uid:           in.Uid,
			UserId:        in.UserId,
			MerchantUid:   in.MerchantUid,
			MerchantId:    in.MerchantId,
			Amount:        fmt.Sprintf("%d", in.Amount),
			State:         consts.C2BBillStateSuccess,
			BizType:       consts.BizTypeC2B,
			Desc:          in.Desc,
			PayTime:       payTime,
		})
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("insert c2b bill failed: %v", err))
		}

		_, err = tc2bPendingTransferModel.Insert(ctx, &mysql.TC2bPendingTransfer{
			TransactionId: in.TransactionId,
			Uid:           in.Uid,
			UserId:        in.UserId,
			MerchantUid:   in.MerchantUid,
			MerchantId:    in.MerchantId,
			Amount:        in.Amount,
			State:         consts.PendingC2bTransferInit,
			Desc:          in.Desc,
		})
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("insert pending c2b transfer failed: %v", err))
		}

		result = &account_mgr_pb.C2BRsp{
			TransactionId: in.TransactionId,
			Uid:           in.Uid,
			UserId:        in.UserId,
			MerchantUid:   in.MerchantUid,
			MerchantId:    in.MerchantId,
			PayTime:       payTime.Format("2006-01-02 15:04:05"),
			IsRepeat:      0,
		}

		return nil
	})
	if err != nil {
		l.Errorf("C2BFinal transaction failed: %v", err)
		return nil, err
	}

	// 异步发送kafka消息, 忽略失败
	go l.sendKafkaMessage(in)

	return result, nil
}

func (l *C2BFinalLogic) sendKafkaMessage(in *account_mgr_pb.C2BReq) {
	message := &PendingC2BTransferMessage{
		TransactionId: in.TransactionId,
	}
	jsonStr, err := json.Marshal(message)
	if err != nil {
		logx.Errorf("marshal message failed: %v", err)
		return
	}
	err = l.svcCtx.C2BAsyncTransferProducer.Push(l.ctx, string(jsonStr))
	// 查看所有topic
	// docker exec -it kafka bash -c "/opt/kafka/bin/kafka-topics.sh --bootstrap-server kafka:29092 --list"
	// 创建topic
	// docker exec -it kafka bash -c "/opt/kafka/bin/kafka-topics.sh --create --bootstrap-server 127.0.0.1:9092 --topic topic_c2b_async_transfer --partitions 2 --replication-factor 1 --config retention.ms=86400000 --config max.message.bytes=10485760"
	// 查看topic详情
	// docker exec -it kafka bash -c "/opt/kafka/bin/kafka-topics.sh --describe --bootstrap-server 127.0.0.1:9092 --topic topic_c2b_async_transfer"
	// 查看topic消息
	// docker exec -it kafka bash -c "/opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server 127.0.0.1:9092 --topic topic_c2b_async_transfer --from-beginning"
	if err != nil {
		logx.Errorf("send kafka message failed: %v", err)
	}
}
