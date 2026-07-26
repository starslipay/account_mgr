package mysql

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ TLocalMessageModel = (*customTLocalMessageModel)(nil)

type (
	// TLocalMessageModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLocalMessageModel.
	TLocalMessageModel interface {
		tLocalMessageModel
		withSession(session sqlx.Session) TLocalMessageModel
	}

	customTLocalMessageModel struct {
		*defaultTLocalMessageModel
	}
)

// NewTLocalMessageModel returns a model for the database table.
func NewTLocalMessageModel(conn sqlx.SqlConn) TLocalMessageModel {
	return &customTLocalMessageModel{
		defaultTLocalMessageModel: newTLocalMessageModel(conn),
	}
}

func (m *customTLocalMessageModel) withSession(session sqlx.Session) TLocalMessageModel {
	return NewTLocalMessageModel(sqlx.NewSqlConnFromSession(session))
}
