package logic

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/account_mgr/internal/svc"
	"github.com/starslipay/account_mgr/internal/xerr"
	"github.com/starslipay/account_mgr/model/mysql"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	QryModeSlave  = 1 // 从库查询
	QryModeMaster = 2 // 主库查询
)

type GetUserBalanceInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserBalanceInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserBalanceInfoLogic {
	return &GetUserBalanceInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserBalanceInfoLogic) GetUserBalanceInfo(in *account_mgr_pb.GetUserBalanceInfoReq) (*account_mgr_pb.GetUserBalanceInfoRsp, error) {
	var model mysql.TCAccountModel
	if QryModeSlave == in.QryMode {
		model = l.svcCtx.TCAccountModelSlave
	} else if QryModeMaster == in.QryMode {
		model = l.svcCtx.TCAccountModelMaster
	} else {
		return nil, xerr.NewParamError("QryMode is invalid, must be 1 or 2")
	}

	account, err := model.FindOne(l.ctx, in.Uid)
	if err != nil {
		return nil, xerr.NewDBError(err.Error())
	}
	return &account_mgr_pb.GetUserBalanceInfoRsp{
		Uid:     account.Uid,
		UserId:  account.UserId,
		Balance: account.Balance,
		CurType: int32(account.CurType),
	}, nil
}
