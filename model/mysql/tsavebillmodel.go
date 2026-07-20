package mysql

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ TSaveBillModel = (*customTSaveBillModel)(nil)

type (
	// TSaveBillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTSaveBillModel.
	TSaveBillModel interface {
		tSaveBillModel
		withSession(session sqlx.Session) TSaveBillModel
	}

	customTSaveBillModel struct {
		*defaultTSaveBillModel
	}
)

// NewTSaveBillModel returns a model for the database table.
func NewTSaveBillModel(conn sqlx.SqlConn) TSaveBillModel {
	return &customTSaveBillModel{
		defaultTSaveBillModel: newTSaveBillModel(conn),
	}
}

func (m *customTSaveBillModel) withSession(session sqlx.Session) TSaveBillModel {
	return NewTSaveBillModel(sqlx.NewSqlConnFromSession(session))
}
