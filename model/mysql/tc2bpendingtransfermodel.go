package mysql

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TC2bPendingTransferModel = (*customTC2bPendingTransferModel)(nil)

type (
	// TC2bPendingTransferModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTC2bPendingTransferModel.
	TC2bPendingTransferModel interface {
		tC2bPendingTransferModel
		withSession(session sqlx.Session) TC2bPendingTransferModel
		FindOneForUpdate(ctx context.Context, transactionId string) (*TC2bPendingTransfer, error)
	}

	customTC2bPendingTransferModel struct {
		*defaultTC2bPendingTransferModel
	}
)

// NewTC2bPendingTransferModel returns a model for the database table.
func NewTC2bPendingTransferModel(conn sqlx.SqlConn) TC2bPendingTransferModel {
	return &customTC2bPendingTransferModel{
		defaultTC2bPendingTransferModel: newTC2bPendingTransferModel(conn),
	}
}

func (m *customTC2bPendingTransferModel) withSession(session sqlx.Session) TC2bPendingTransferModel {
	return NewTC2bPendingTransferModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customTC2bPendingTransferModel) FindOneForUpdate(ctx context.Context, transactionId string) (*TC2bPendingTransfer, error) {
	query := fmt.Sprintf("select %s from %s where `transaction_id` = ? for update", tC2bPendingTransferRows, m.table)
	var resp TC2bPendingTransfer
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