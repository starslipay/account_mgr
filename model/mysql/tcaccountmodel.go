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
		SubBalance(ctx context.Context, uid int64, amount int64) error
		FindOneForUpdate(ctx context.Context, uid int64) (*TCAccount, error)
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

func (m *customTCAccountModel) SubBalance(ctx context.Context, uid int64, amount int64) error {
	query := fmt.Sprintf("update %s set `balance` = `balance` - ? where `uid` = ? and `balance` >= ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, amount, uid, amount)
	return err
}

func (m *customTCAccountModel) FindOneForUpdate(ctx context.Context, uid int64) (*TCAccount, error) {
	query := fmt.Sprintf("select %s from %s where `uid` = ? for update", tCAccountRows, m.table)
	var resp TCAccount
	err := m.conn.QueryRowCtx(ctx, &resp, query, uid)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
