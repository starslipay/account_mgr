package mysql

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TBAccountModel = (*customTBAccountModel)(nil)

type (
	// TBAccountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTBAccountModel.
	TBAccountModel interface {
		tBAccountModel
		withSession(session sqlx.Session) TBAccountModel
		AddBalance(ctx context.Context, merchantUid int64, amount int64) error
		FindOneForUpdate(ctx context.Context, merchantUid int64) (*TBAccount, error)
	}

	customTBAccountModel struct {
		*defaultTBAccountModel
	}
)

// NewTBAccountModel returns a model for the database table.
func NewTBAccountModel(conn sqlx.SqlConn) TBAccountModel {
	return &customTBAccountModel{
		defaultTBAccountModel: newTBAccountModel(conn),
	}
}

func (m *customTBAccountModel) withSession(session sqlx.Session) TBAccountModel {
	return NewTBAccountModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customTBAccountModel) AddBalance(ctx context.Context, merchantUid int64, amount int64) error {
	query := fmt.Sprintf("update %s set `balance` = `balance` + ? where `merchant_uid` = ?", m.table)
	ret, err := m.conn.ExecCtx(ctx, query, amount, merchantUid)
	if err != nil {
		return err
	}
	return checkOneRowAffected(ret)
}

func (m *customTBAccountModel) FindOneForUpdate(ctx context.Context, merchantUid int64) (*TBAccount, error) {
	query := fmt.Sprintf("select %s from %s where `merchant_uid` = ? for update", tBAccountRows, m.table)
	var resp TBAccount
	err := m.conn.QueryRowCtx(ctx, &resp, query, merchantUid)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
