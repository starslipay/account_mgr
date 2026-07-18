package logic

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/account_mgr/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
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

	return &account_mgr_pb.Bank2CRsp{}, nil
}
