package mysql

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ TC2bBillModel = (*customTC2bBillModel)(nil)

type (
	// TC2bBillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTC2bBillModel.
	TC2bBillModel interface {
		tC2bBillModel
		withSession(session sqlx.Session) TC2bBillModel
	}

	customTC2bBillModel struct {
		*defaultTC2bBillModel
	}
)

// NewTC2bBillModel returns a model for the database table.
func NewTC2bBillModel(conn sqlx.SqlConn) TC2bBillModel {
	return &customTC2bBillModel{
		defaultTC2bBillModel: newTC2bBillModel(conn),
	}
}

func (m *customTC2bBillModel) withSession(session sqlx.Session) TC2bBillModel {
	return NewTC2bBillModel(sqlx.NewSqlConnFromSession(session))
}
