package mysql

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TCAccountLogModel = (*customTCAccountLogModel)(nil)

type (
	// TCAccountLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTCAccountLogModel.
	TCAccountLogModel interface {
		tCAccountLogModel
		withSession(session sqlx.Session) TCAccountLogModel
		ListByUid(ctx context.Context, uid int64, offset, limit int) ([]*TCAccountLog, error)
	}

	customTCAccountLogModel struct {
		*defaultTCAccountLogModel
	}
)

// NewTCAccountLogModel returns a model for the database table.
func NewTCAccountLogModel(conn sqlx.SqlConn) TCAccountLogModel {
	return &customTCAccountLogModel{
		defaultTCAccountLogModel: newTCAccountLogModel(conn),
	}
}

func (m *customTCAccountLogModel) withSession(session sqlx.Session) TCAccountLogModel {
	return NewTCAccountLogModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customTCAccountLogModel) ListByUid(ctx context.Context, uid int64, offset, limit int) ([]*TCAccountLog, error) {
	query := fmt.Sprintf("select %s from %s where `uid` = ? order by `create_time` desc limit ? offset ?", tCAccountLogRows, m.table)
	var resp []*TCAccountLog
	err := m.conn.QueryRowsCtx(ctx, &resp, query, uid, limit, offset)
	return resp, err
}
