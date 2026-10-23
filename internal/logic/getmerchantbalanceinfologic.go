package logic

import (
	"context"
	"errors"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/account_mgr/internal/svc"
	"github.com/starslipay/account_mgr/internal/xerr"
	"github.com/starslipay/account_mgr/model/mysql"
	"github.com/starslipay/paycomm/xerror"
	"google.golang.org/grpc/codes"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMerchantBalanceInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMerchantBalanceInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMerchantBalanceInfoLogic {
	return &GetMerchantBalanceInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMerchantBalanceInfoLogic) GetMerchantBalanceInfo(in *account_mgr_pb.GetMerchantBalanceInfoReq) (*account_mgr_pb.GetMerchantBalanceInfoRsp, error) {
	if in.MerchantUid <= 0 {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "merchant_uid is invalid")
	}
	if in.QryMode != 1 && in.QryMode != 2 {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "qry_mode is invalid")
	}

	var (
		account *mysql.TBAccount
		err     error
	)
	if in.QryMode == 1 {
		account, err = l.svcCtx.TBAccountModelSlave.FindOne(l.ctx, in.MerchantUid)
	} else {
		account, err = l.svcCtx.TBAccountModelMaster.FindOne(l.ctx, in.MerchantUid)
	}
	if err != nil {
		if errors.Is(err, mysql.ErrNotFound) {
			return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeMerchantNotFound, "merchant account not found")
		}
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, "find merchant account failed")
	}

	return &account_mgr_pb.GetMerchantBalanceInfoRsp{
		MerchantUid: account.MerchantUid,
		MerchantId:  account.MerchantId,
		Balance:     account.Balance,
		CurType:     int32(account.CurType),
	}, nil
}
