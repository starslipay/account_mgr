package logic

import (
	"context"
	"fmt"

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

type C2CAsyncAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewC2CAsyncAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *C2CAsyncAccountLogic {
	return &C2CAsyncAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *C2CAsyncAccountLogic) C2CAsyncAccount(in *account_mgr_pb.C2CAsyncAccountReq) (*account_mgr_pb.C2CAsyncAccountRsp, error) {
	// 1. 先无锁查询t_pending_c2c_transfer，查看是否已经入账（幂等检查）
	pendingTransfer, err := l.svcCtx.TLocalPendingC2cTransferModelSlave.FindOne(l.ctx, in.TransactionId)
	if err != nil {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("find pending c2c transfer failed: %v", err))
	}

	// 已入账，直接返回成功（幂等）
	if pendingTransfer.State == consts.PendingC2cTransferDone {
		return &account_mgr_pb.C2CAsyncAccountRsp{}, nil
	}

	// 2. 加锁查询并执行入账逻辑
	err = l.svcCtx.SqlMasterConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		tcAccountModel := mysql.NewTCAccountModel(sqlx.NewSqlConnFromSession(session))
		tcAccountLogModel := mysql.NewTCAccountLogModel(sqlx.NewSqlConnFromSession(session))
		tPendingC2cTransferModel := mysql.NewTPendingC2cTransferModel(sqlx.NewSqlConnFromSession(session))

		// 加锁查询t_pending_c2c_transfer
		pendingTransfer, err := tPendingC2cTransferModel.FindOneForUpdate(ctx, in.TransactionId)
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("find pending c2c transfer for update failed: %v", err))
		}

		// 二次检查是否已入账（防止并发重复入账）
		if pendingTransfer.State == consts.PendingC2cTransferDone {
			return nil
		}

		sellerAccount, err := tcAccountModel.FindOneForUpdate(ctx, pendingTransfer.SellerUid)
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("find seller account failed: %v", err))
		}

		// 卖家账户加钱
		err = tcAccountModel.AddBalance(ctx, pendingTransfer.SellerUid, pendingTransfer.Amount)
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("add balance failed: %v", err))
		}

		// 记录卖家入账流水
		_, err = tcAccountLogModel.Insert(ctx, &mysql.TCAccountLog{
			Uid:                pendingTransfer.SellerUid,
			UserId:             pendingTransfer.SellerUserId,
			CounterpartyUserId: pendingTransfer.BuyerUserId,
			CounterpartyUid:    pendingTransfer.BuyerUid,
			TransactionId:      pendingTransfer.TransactionId,
			InoutType:          consts.InoutTypeIn,
			BizType:            consts.BizTypeC2C,
			Balance:            sellerAccount.Balance + pendingTransfer.Amount,
			Amount:             pendingTransfer.Amount,
			Desc:               pendingTransfer.Desc,
		})
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("insert account log failed: %v", err))
		}

		// 修改t_pending_c2c_transfer状态为已完成
		pendingTransfer.State = consts.PendingC2cTransferDone
		err = tPendingC2cTransferModel.Update(ctx, pendingTransfer)
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("update pending c2c transfer state failed: %v", err))
		}

		return nil
	})
	if err != nil {
		l.Errorf("C2CAsyncAccount transaction failed: %v", err)
		return nil, err
	}

	return &account_mgr_pb.C2CAsyncAccountRsp{}, nil
}
