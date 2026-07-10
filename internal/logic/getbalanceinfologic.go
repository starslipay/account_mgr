package logic

import (
	"context"
	"errors"

	"account_mgr/account_mgr_pb"
	"account_mgr/internal/svc"
	"account_mgr/model/mysql"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	QryModeSlave  = 1 // 从库查询
	QryModeMaster = 2 // 主库查询
)

type GetBalanceInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBalanceInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBalanceInfoLogic {
	return &GetBalanceInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetBalanceInfoLogic) GetBalanceInfo(in *account_mgr_pb.GetBalanceInfoReq) (*account_mgr_pb.GetBalanceInfoRsp, error) {
	var model mysql.TCAccountModel
	if QryModeSlave == in.QryMode {
		model = l.svcCtx.TCAccountModelSlave
	} else if QryModeMaster == in.QryMode {
		model = l.svcCtx.TCAccountModelMaster
	} else {
		return nil, errors.New("QryMode is invalid, must be 1 or 2")
	}

	account, err := model.FindOne(l.ctx, in.Uid)
	if err != nil {
		return nil, err
	}
	return &account_mgr_pb.GetBalanceInfoRsp{
		Uid:     account.Uid,
		UserId:  account.UserId,
		Balance: account.Balance,
		CurType: int32(account.CurType),
	}, nil
}
