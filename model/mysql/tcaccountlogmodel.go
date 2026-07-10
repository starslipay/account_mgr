package mysql

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ TCAccountLogModel = (*customTCAccountLogModel)(nil)

type (
	// TCAccountLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTCAccountLogModel.
	TCAccountLogModel interface {
		tCAccountLogModel
		withSession(session sqlx.Session) TCAccountLogModel
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
