package logic

import (
	"context"
	"fmt"
	"strconv"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/account_mgr/internal/svc"
	"github.com/starslipay/account_mgr/internal/xerr"
	"github.com/starslipay/account_mgr/model/mysql"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	InoutTypeIn     = 1 // 资金方向：入
	InoutTypeOut    = 2 // 资金方向：出
	BizTypeBank2C   = 1 // 银行充值
	SaveBillStateOK = 1 // 充值单状态：成功
)

type Bank2CLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBank2CLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Bank2CLogic {
	return &Bank2CLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *Bank2CLogic) Bank2C(in *account_mgr_pb.Bank2CReq) (*account_mgr_pb.Bank2CRsp, error) {
	if in.Uid <= 0 {
		return nil, xerr.NewParamError("uid is invalid")
	}
	if in.Amount <= 0 {
		return nil, xerr.NewParamError("amount must be positive")
	}
	if in.TransactionId == "" {
		return nil, xerr.NewParamError("transaction_id is required")
	}

	err := l.svcCtx.SqlMasterConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		tCAccountModel := mysql.NewTCAccountModel(sqlx.NewSqlConnFromSession(session))
		tCAccountLogModel := mysql.NewTCAccountLogModel(sqlx.NewSqlConnFromSession(session))
		tSaveBillModel := mysql.NewTSaveBillModel(sqlx.NewSqlConnFromSession(session))

		err := tCAccountModel.AddBalance(ctx, in.Uid, in.Amount)
		if err != nil {
			return err
		}

		_, err = tCAccountLogModel.Insert(ctx, &mysql.TCAccountLog{
			Uid:                in.Uid,
			UserId:             in.UserId,
			CounterpartyUserId: strconv.Itoa(int(in.BankType)),
			CounterpartyUid:    int64(in.BankType),
			TransactionId:      in.TransactionId,
			InoutType:          InoutTypeIn,
			BizType:            BizTypeBank2C,
			Amount:             in.Amount,
			Desc:               in.Desc,
		})
		if err != nil {
			return err
		}

		_, err = tSaveBillModel.Insert(ctx, &mysql.TSaveBill{
			TransactionId: in.TransactionId,
			Uid:           in.Uid,
			UserId:        in.UserId,
			BankType:      strconv.Itoa(int(in.BankType)),
			Amount:        in.Amount,
			State:         SaveBillStateOK,
			Desc:          in.Desc,
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		l.Errorf("Bank2C transaction failed: %v", err)
		return nil, xerr.NewDBError(fmt.Sprintf("transaction failed: %v", err))
	}

	return &account_mgr_pb.Bank2CRsp{
		TransactionId: in.TransactionId,
		UserId:        in.UserId,
	}, nil
}
