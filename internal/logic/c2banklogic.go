package logic

import (
	"context"
	"fmt"
	"strconv"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/account_mgr/internal/svc"
	"github.com/starslipay/account_mgr/internal/xerr"
	"github.com/starslipay/account_mgr/model/mysql"
	"github.com/starslipay/paycomm/xerror"
	"google.golang.org/grpc/codes"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type C2BankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewC2BankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *C2BankLogic {
	return &C2BankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *C2BankLogic) C2Bank(in *account_mgr_pb.C2BankReq) (*account_mgr_pb.C2BankRsp, error) {
	if in.Uid <= 0 {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "uid is invalid")
	}
	if in.Amount <= 0 {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "amount must be positive")
	}
	if in.TransactionId == "" {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "transaction_id is required")
	}

	err := l.svcCtx.SqlMasterConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		tCAccountModel := mysql.NewTCAccountModel(sqlx.NewSqlConnFromSession(session))
		tCAccountLogModel := mysql.NewTCAccountLogModel(sqlx.NewSqlConnFromSession(session))
		tSaveBillModel := mysql.NewTSaveBillModel(sqlx.NewSqlConnFromSession(session))

		account, err := tCAccountModel.FindOneForUpdate(ctx, in.Uid)
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("find account failed: %v", err))
		}

		if account.Balance < in.Amount {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeBalanceNotEnough, "balance is not enough")
		}

		err = tCAccountModel.SubBalance(ctx, in.Uid, in.Amount)
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("add balance failed: %v", err))
		}

		_, err = tCAccountLogModel.Insert(ctx, &mysql.TCAccountLog{
			Uid:                in.Uid,
			UserId:             in.UserId,
			CounterpartyUserId: strconv.Itoa(int(in.BankType)),
			CounterpartyUid:    int64(in.BankType),
			TransactionId:      in.TransactionId,
			InoutType:          InoutTypeOut,
			BizType:            BizTypeC2Bank,
			Amount:             in.Amount,
			Desc:               in.Desc,
		})
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("insert account log failed: %v", err))
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
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("insert save bill failed: %v", err))
		}

		return nil
	})

	if err != nil {
		l.Errorf("C2Bank transaction failed: %v", err)
		return nil, err
	}

	return &account_mgr_pb.C2BankRsp{
		TransactionId: in.TransactionId,
		UserId:        in.UserId,
	}, nil
}
