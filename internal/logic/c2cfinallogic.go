package logic

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/account_mgr/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type C2cFinalLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewC2cFinalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *C2cFinalLogic {
	return &C2cFinalLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *C2cFinalLogic) C2CFinal(in *account_mgr_pb.C2CReq) (*account_mgr_pb.C2CRsp, error) {
	return &account_mgr_pb.C2CRsp{}, nil
}
