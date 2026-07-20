package mysql

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TCAccountModel = (*customTCAccountModel)(nil)

type (
	// TCAccountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTCAccountModel.
	TCAccountModel interface {
		tCAccountModel
		withSession(session sqlx.Session) TCAccountModel
		AddBalance(ctx context.Context, uid int64, amount int64) error
	}

	customTCAccountModel struct {
		*defaultTCAccountModel
	}
)

// NewTCAccountModel returns a model for the database table.
func NewTCAccountModel(conn sqlx.SqlConn) TCAccountModel {
	return &customTCAccountModel{
		defaultTCAccountModel: newTCAccountModel(conn),
	}
}

func (m *customTCAccountModel) withSession(session sqlx.Session) TCAccountModel {
	return NewTCAccountModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customTCAccountModel) AddBalance(ctx context.Context, uid int64, amount int64) error {
	query := fmt.Sprintf("update %s set `balance` = `balance` + ? where `uid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, amount, uid)
	return err
}
