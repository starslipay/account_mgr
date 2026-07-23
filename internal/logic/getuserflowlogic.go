package logic

import (
	"context"
	"fmt"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/account_mgr/internal/svc"
	"github.com/starslipay/account_mgr/internal/xerr"
	"github.com/starslipay/paycomm/xerror"
	"google.golang.org/grpc/codes"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserFlowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserFlowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserFlowLogic {
	return &GetUserFlowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserFlowLogic) GetUserFlow(in *account_mgr_pb.GetUserFlowReq) (*account_mgr_pb.GetUserFlowRsp, error) {
	if in.Uid <= 0 {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParam, "uid is invalid")
	}
	if in.Limit <= 0 {
		in.Limit = 20
	}

	account, err := l.svcCtx.TCAccountModelSlave.FindOne(l.ctx, in.Uid)
	if err != nil {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("find account failed: %v", err))
	}

	// 为了判断是否还有下一页，所以这里查询的limit+1
	flows, err := l.svcCtx.TCAccountLogModelSlave.ListByUid(l.ctx, in.Uid, int(in.Offset), int(in.Limit)+1)
	if err != nil {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeDB, fmt.Sprintf("list flow failed: %v", err))
	}

	var userFlowList []*account_mgr_pb.UserFlow
	for i, flow := range flows {
		if i >= int(in.Limit) {
			break
		}
		userFlowList = append(userFlowList, &account_mgr_pb.UserFlow{
			TransactionId:      flow.TransactionId,
			UserId:             flow.UserId,
			CounterpartyUserId: flow.CounterpartyUserId,
			InoutType:          int32(flow.InoutType),
			BizType:            int32(flow.BizType),
			Amount:             flow.Amount,
			Balance:            flow.Balance,
			Desc:               flow.Desc,
			CreateTime:         flow.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}

	var nextOffset int32
	var endFlag int32
	if len(flows) > int(in.Limit) {
		endFlag = 0
		nextOffset = in.Offset + in.Limit
	} else {
		endFlag = 1
		nextOffset = 0
	}

	return &account_mgr_pb.GetUserFlowRsp{
		Uid:          in.Uid,
		UserId:       account.UserId,
		NextOffset:   nextOffset,
		EndFlag:      endFlag,
		UserFlowList: userFlowList,
	}, nil
}
