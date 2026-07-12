package logic

import (
	"context"
	"math"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/account_mgr/internal/svc"
	"github.com/starslipay/account_mgr/internal/xerr"
	"github.com/starslipay/account_mgr/model/mysql"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAccountLogic {
	return &CreateAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateAccountLogic) CreateAccount(in *account_mgr_pb.CreateAccountReq) (*account_mgr_pb.CreateAccountRsp, error) {
	if in.Uid < 1 || in.Uid > (math.MaxInt64-1) {
		return nil, xerr.NewParamError("Uid is invalid, must be [1, int64.max-1]")
	}

	isAccountExist := true
	account, err := l.svcCtx.TCAccountModelMaster.FindOne(l.ctx, in.Uid)
	if err != nil {
		if err == mysql.ErrNotFound {
			isAccountExist = false
		} else {
			return nil, xerr.NewDBError(err.Error())
		}
	}
	if isAccountExist {
		// 判断重入前，校验关键字段一致性
		if account.UserId != in.UserId {
			return nil, xerr.NewParamError("user id not match")
		}

		if account.CurType != int64(in.CurType) {
			return nil, xerr.NewParamError("cur type not match")
		}

		return &account_mgr_pb.CreateAccountRsp{
			Uid:      account.Uid,
			UserId:   account.UserId,
			IsRepeat: true,
		}, nil
	}

	_, err = l.svcCtx.TCAccountModelMaster.Insert(l.ctx, &mysql.TCAccount{
		Uid:     in.Uid,
		UserId:  in.UserId,
		CurType: int64(in.CurType),
		Balance: 0,
	})
	if err != nil {
		return nil, xerr.NewDBError(err.Error())
	}

	return &account_mgr_pb.CreateAccountRsp{
		Uid:      in.Uid,
		UserId:   in.UserId,
		IsRepeat: false,
	}, nil
}
