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

type C2BAsyncAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewC2BAsyncAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *C2BAsyncAccountLogic {
	return &C2BAsyncAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *C2BAsyncAccountLogic) C2BAsyncAccount(in *account_mgr_pb.C2BAsyncAccountReq) (*account_mgr_pb.C2BAsyncAccountRsp, error) {
	// 1. 先无锁查询t_c2b_pending_transfer，查看是否已经入账（幂等检查）
	pendingTransfer, err := l.svcCtx.TC2bPendingTransferModelMaster.FindOne(l.ctx, in.TransactionId)
	if err != nil {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("find pending c2b transfer failed: %v", err))
	}

	// 已入账，直接返回成功（幂等）
	if pendingTransfer.State == consts.PendingC2bTransferDone {
		return &account_mgr_pb.C2BAsyncAccountRsp{}, nil
	}

	// 2. 加锁查询并执行入账逻辑
	err = l.svcCtx.SqlMasterConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		tbAccountModel := mysql.NewTBAccountModel(sqlx.NewSqlConnFromSession(session))
		tbAccountLogModel := mysql.NewTBAccountLogModel(sqlx.NewSqlConnFromSession(session))
		tC2bPendingTransferModel := mysql.NewTC2bPendingTransferModel(sqlx.NewSqlConnFromSession(session))

		// 加锁查询t_c2b_pending_transfer
		pendingTransfer, err := tC2bPendingTransferModel.FindOneForUpdate(ctx, in.TransactionId)
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("find pending c2b transfer for update failed: %v", err))
		}

		// 二次检查是否已入账（防止并发重复入账）
		if pendingTransfer.State == consts.PendingC2bTransferDone {
			return nil
		}

		merchantAccount, err := tbAccountModel.FindOneForUpdate(ctx, pendingTransfer.MerchantUid)
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("find merchant account failed: %v", err))
		}

		// 商户账户加钱
		err = tbAccountModel.AddBalance(ctx, pendingTransfer.MerchantUid, pendingTransfer.Amount)
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("add balance failed: %v", err))
		}

		// 记录商户入账流水
		_, err = tbAccountLogModel.Insert(ctx, &mysql.TBAccountLog{
			MerchantUid:     pendingTransfer.MerchantUid,
			MerchantId:      pendingTransfer.MerchantId,
			CounterpartyId:  pendingTransfer.UserId,
			CounterpartyUid: pendingTransfer.Uid,
			TransactionId:   pendingTransfer.TransactionId,
			InoutType:       consts.InoutTypeIn,
			BizType:         consts.BizTypeC2B,
			Balance:         merchantAccount.Balance + pendingTransfer.Amount,
			Amount:          pendingTransfer.Amount,
			Desc:            pendingTransfer.Desc,
		})
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("insert account log failed: %v", err))
		}

		// 修改t_c2b_pending_transfer状态为已完成
		pendingTransfer.State = consts.PendingC2bTransferDone
		err = tC2bPendingTransferModel.Update(ctx, pendingTransfer)
		if err != nil {
			return xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("update pending c2b transfer state failed: %v", err))
		}

		return nil
	})
	if err != nil {
		l.Errorf("C2BAsyncAccount transaction failed: %v", err)
		return nil, err
	}

	return &account_mgr_pb.C2BAsyncAccountRsp{}, nil
}
