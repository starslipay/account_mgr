package mysql

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ TC2cBillModel = (*customTC2cBillModel)(nil)

type (
	// TC2cBillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTC2cBillModel.
	TC2cBillModel interface {
		tC2cBillModel
		withSession(session sqlx.Session) TC2cBillModel
	}

	customTC2cBillModel struct {
		*defaultTC2cBillModel
	}
)

// NewTC2cBillModel returns a model for the database table.
func NewTC2cBillModel(conn sqlx.SqlConn) TC2cBillModel {
	return &customTC2cBillModel{
		defaultTC2cBillModel: newTC2cBillModel(conn),
	}
}

func (m *customTC2cBillModel) withSession(session sqlx.Session) TC2cBillModel {
	return NewTC2cBillModel(sqlx.NewSqlConnFromSession(session))
}
