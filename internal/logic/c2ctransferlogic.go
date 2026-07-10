package logic

import (
	"context"

	"account_mgr/account_mgr_pb"
	"account_mgr/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type C2cTransferLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewC2cTransferLogic(ctx context.Context, svcCtx *svc.ServiceContext) *C2cTransferLogic {
	return &C2cTransferLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *C2cTransferLogic) C2CTransfer(in *account_mgr_pb.C2CTransferReq) (*account_mgr_pb.C2CTransferRsp, error) {
	// todo: add your logic here and delete this line

	return &account_mgr_pb.C2CTransferRsp{}, nil
}
