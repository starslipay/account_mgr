package mysql

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ TBAccountModel = (*customTBAccountModel)(nil)

type (
	// TBAccountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTBAccountModel.
	TBAccountModel interface {
		tBAccountModel
		withSession(session sqlx.Session) TBAccountModel
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
