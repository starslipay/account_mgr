package logic

import (
	"context"
	"fmt"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/account_mgr/internal/svc"
	"github.com/starslipay/account_mgr/internal/xerr"
	"github.com/starslipay/paycomm/xerror"
	"google.golang.org/grpc/codes"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryC2BBillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryC2BBillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryC2BBillLogic {
	return &QueryC2BBillLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryC2BBillLogic) QueryC2BBill(in *account_mgr_pb.QueryC2BBillReq) (*account_mgr_pb.QueryC2BBillRsp, error) {
	if in.TransactionId == "" {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "transaction_id is invalid")
	}

	bill, err := l.svcCtx.TC2bBillModelMaster.FindOne(l.ctx, in.TransactionId)
	if err != nil {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("find bill failed: %v", err))
	}

	return &account_mgr_pb.QueryC2BBillRsp{
		TransactionId: bill.TransactionId,
		OutTradeNo:    bill.OutTradeNo,
		UserId:        bill.UserId,
		Uid:           bill.Uid,
		MerchantUid:   bill.MerchantUid,
		MerchantId:    bill.MerchantId,
		Amount:        bill.Amount,
		State:         int32(bill.State),
	}, nil
}
