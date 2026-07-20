package logic

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/account_mgr/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type C2cLocalLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewC2cLocalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *C2cLocalLogic {
	return &C2cLocalLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *C2cLocalLogic) C2CLocal(in *account_mgr_pb.C2CReq) (*account_mgr_pb.C2CRsp, error) {
	// todo: add your logic here and delete this line

	return &account_mgr_pb.C2CRsp{}, nil
}
