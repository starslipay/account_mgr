package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/account_mgr/internal/svc"
	"github.com/starslipay/account_mgr/internal/xerr"
	"github.com/starslipay/account_mgr/model/mysql"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	C2cBillStateOK  = 1
	BizTypeC2cLocal = 2
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
	if in.BuyerUid <= 0 {
		return nil, xerr.NewParamError("buyer_uid is invalid")
	}
	if in.SellerUid <= 0 {
		return nil, xerr.NewParamError("seller_uid is invalid")
	}
	if in.BuyerUid == in.SellerUid {
		return nil, xerr.NewParamError("buyer and seller cannot be the same")
	}
	if in.Amount <= 0 {
		return nil, xerr.NewParamError("amount must be positive")
	}
	if in.TransactionId == "" {
		return nil, xerr.NewParamError("transaction_id is required")
	}

	var result *account_mgr_pb.C2CRsp
	err := l.svcCtx.SqlMasterConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		tcAccountModel := mysql.NewTCAccountModel(sqlx.NewSqlConnFromSession(session))
		tcAccountLogModel := mysql.NewTCAccountLogModel(sqlx.NewSqlConnFromSession(session))
		tc2cBillModel := mysql.NewTC2cBillModel(sqlx.NewSqlConnFromSession(session))

		var buyerAccount *mysql.TCAccount
		var err error
		if in.BuyerUid < in.SellerUid {
			buyerAccount, err = tcAccountModel.FindOneForUpdate(ctx, in.BuyerUid)
			if err != nil {
				return err
			}
			_, err = tcAccountModel.FindOneForUpdate(ctx, in.SellerUid)
			if err != nil {
				return err
			}
		} else {
			_, err = tcAccountModel.FindOneForUpdate(ctx, in.SellerUid)
			if err != nil {
				return err
			}
			buyerAccount, err = tcAccountModel.FindOneForUpdate(ctx, in.BuyerUid)
			if err != nil {
				return err
			}
		}

		_ = buyerAccount
		if buyerAccount.Balance < in.Amount {
			return xerr.ErrBalanceNotEnough
		}

		err = tcAccountModel.SubBalance(ctx, in.BuyerUid, in.Amount)
		if err != nil {
			return err
		}

		err = tcAccountModel.AddBalance(ctx, in.SellerUid, in.Amount)
		if err != nil {
			return err
		}

		_, err = tcAccountLogModel.Insert(ctx, &mysql.TCAccountLog{
			Uid:                in.BuyerUid,
			UserId:             in.BuyerUserId,
			CounterpartyUserId: in.SellerUserId,
			CounterpartyUid:    in.SellerUid,
			TransactionId:      in.TransactionId,
			InoutType:          InoutTypeIn,
			BizType:            BizTypeC2cLocal,
			Amount:             in.Amount,
		})
		if err != nil {
			return err
		}

		_, err = tcAccountLogModel.Insert(ctx, &mysql.TCAccountLog{
			Uid:                in.SellerUid,
			UserId:             in.SellerUserId,
			CounterpartyUserId: in.BuyerUserId,
			CounterpartyUid:    in.BuyerUid,
			TransactionId:      in.TransactionId,
			InoutType:          InoutTypeOut,
			BizType:            BizTypeC2cLocal,
			Amount:             in.Amount,
		})
		if err != nil {
			return err
		}

		_, err = tc2cBillModel.Insert(ctx, &mysql.TC2cBill{
			TransactionId: in.TransactionId,
			BuyerUid:      in.BuyerUid,
			SellerUid:     in.SellerUid,
			BuyerUserId:   in.BuyerUserId,
			SellerUserId:  in.SellerUserId,
			Amount:        in.Amount,
			State:         C2cBillStateOK,
			BizType:       BizTypeC2cLocal,
			Desc:          in.Desc,
		})
		if err != nil {
			return err
		}

		result = &account_mgr_pb.C2CRsp{
			TransactionId: in.TransactionId,
			BuyerUid:      in.BuyerUid,
			BuyerUserId:   in.BuyerUserId,
			SellerUid:     in.SellerUid,
			SellerUserId:  in.SellerUserId,
			TransferTime:  time.Now().Format("2006-01-02 15:04:05"),
			IsRepeat:      0,
		}

		return nil
	})

	if err != nil {
		l.Errorf("C2CLocal transaction failed: %v", err)
		return nil, xerr.NewDBError(fmt.Sprintf("transaction failed: %v", err))
	}

	return result, nil
}
