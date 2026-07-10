package mysql

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ TCAccountModel = (*customTCAccountModel)(nil)

type (
	// TCAccountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTCAccountModel.
	TCAccountModel interface {
		tCAccountModel
		withSession(session sqlx.Session) TCAccountModel
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
