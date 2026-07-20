package logic

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/account_mgr/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type C2cStrongLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewC2cStrongLogic(ctx context.Context, svcCtx *svc.ServiceContext) *C2cStrongLogic {
	return &C2cStrongLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *C2cStrongLogic) C2CStrong(in *account_mgr_pb.C2CReq) (*account_mgr_pb.C2CRsp, error) {
	// todo: add your logic here and delete this line

	return &account_mgr_pb.C2CRsp{}, nil
}
