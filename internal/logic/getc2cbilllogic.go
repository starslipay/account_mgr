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
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type GetC2CBillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetC2CBillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetC2CBillLogic {
	return &GetC2CBillLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetC2CBillLogic) GetC2CBill(in *account_mgr_pb.GetC2CBillReq) (*account_mgr_pb.GetC2CBillRsp, error) {
	bill, err := l.svcCtx.TC2cBillModelSlave.FindOne(l.ctx, in.TransactionId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeC2CBillNotFound, "bill not found")
		}
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("find bill failed: %v", err))
	}
	return &account_mgr_pb.GetC2CBillRsp{
		TransactionId: bill.TransactionId,
		BuyerUid:      bill.BuyerUid,
		BuyerUserId:   bill.BuyerUserId,
		SellerUid:     bill.SellerUid,
		SellerUserId:  bill.SellerUserId,
		PayTime:       bill.PayTime.Format("2006-01-02 15:04:05"),
		Amount:        bill.Amount,
		Desc:          bill.Desc,
	}, nil
}
