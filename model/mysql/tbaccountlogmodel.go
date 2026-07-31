package mysql

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ TBAccountLogModel = (*customTBAccountLogModel)(nil)

type (
	// TBAccountLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTBAccountLogModel.
	TBAccountLogModel interface {
		tBAccountLogModel
		withSession(session sqlx.Session) TBAccountLogModel
	}

	customTBAccountLogModel struct {
		*defaultTBAccountLogModel
	}
)

// NewTBAccountLogModel returns a model for the database table.
func NewTBAccountLogModel(conn sqlx.SqlConn) TBAccountLogModel {
	return &customTBAccountLogModel{
		defaultTBAccountLogModel: newTBAccountLogModel(conn),
	}
}

func (m *customTBAccountLogModel) withSession(session sqlx.Session) TBAccountLogModel {
	return NewTBAccountLogModel(sqlx.NewSqlConnFromSession(session))
}
