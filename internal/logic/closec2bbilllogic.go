package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	driverMysql "github.com/go-sql-driver/mysql"
	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/account_mgr/internal/consts"
	"github.com/starslipay/account_mgr/internal/svc"
	"github.com/starslipay/account_mgr/internal/xerr"
	"github.com/starslipay/account_mgr/model/mysql"
	"github.com/starslipay/paycomm/xerror"
	"google.golang.org/grpc/codes"

	"github.com/zeromicro/go-zero/core/logx"
)

// mysqlErrDuplicateEntry MySQL唯一键冲突错误码
const mysqlErrDuplicateEntry = 1062

type CloseC2BBillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCloseC2BBillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CloseC2BBillLogic {
	return &CloseC2BBillLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CloseC2BBillLogic) CloseC2BBill(in *account_mgr_pb.CloseC2BBillReq) (*account_mgr_pb.CloseC2BBillRsp, error) {
	if in.TransactionId == "" {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "transaction_id is required")
	}

	// 直接插入一条状态为关闭的C2B单据, transaction_id为主键, 冲突则说明单据已存在
	_, err := l.svcCtx.TC2bBillModelMaster.Insert(l.ctx, &mysql.TC2bBill{
		TransactionId: in.TransactionId,
		State:         consts.C2BBillStateClose,
		BizType:       consts.BizTypeC2B,
		PayTime:       time.Now(),
	})
	if err != nil {
		var mysqlErr *driverMysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlErrDuplicateEntry {
			return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeC2BBillConflict, "c2b bill already exists")
		}
		l.Errorf("CloseC2BBill insert failed: %v", err)
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("insert c2b bill failed: %v", err))
	}

	return &account_mgr_pb.CloseC2BBillRsp{
		TransactionId: in.TransactionId,
	}, nil
}
