package mysql

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TC2cPendingTransferModel = (*customTC2cPendingTransferModel)(nil)

type (
	// TC2cPendingTransferModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTC2cPendingTransferModel.
	TC2cPendingTransferModel interface {
		tC2cPendingTransferModel
		withSession(session sqlx.Session) TC2cPendingTransferModel
		FindOneForUpdate(ctx context.Context, transactionId string) (*TC2cPendingTransfer, error)
	}

	customTC2cPendingTransferModel struct {
		*defaultTC2cPendingTransferModel
	}
)

// NewTC2cPendingTransferModel returns a model for the database table.
func NewTC2cPendingTransferModel(conn sqlx.SqlConn) TC2cPendingTransferModel {
	return &customTC2cPendingTransferModel{
		defaultTC2cPendingTransferModel: newTC2cPendingTransferModel(conn),
	}
}

func (m *customTC2cPendingTransferModel) withSession(session sqlx.Session) TC2cPendingTransferModel {
	return NewTC2cPendingTransferModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customTC2cPendingTransferModel) FindOneForUpdate(ctx context.Context, transactionId string) (*TC2cPendingTransfer, error) {
	query := fmt.Sprintf("select %s from %s where `transaction_id` = ? for update", tC2cPendingTransferRows, m.table)
	var resp TC2cPendingTransfer
	err := m.conn.QueryRowCtx(ctx, &resp, query, transactionId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
