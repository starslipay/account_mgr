package mysql

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ TC2cOrderModel = (*customTC2cOrderModel)(nil)

type (
	// TC2cOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTC2cOrderModel.
	TC2cOrderModel interface {
		tC2cOrderModel
		withSession(session sqlx.Session) TC2cOrderModel
	}

	customTC2cOrderModel struct {
		*defaultTC2cOrderModel
	}
)

// NewTC2cOrderModel returns a model for the database table.
func NewTC2cOrderModel(conn sqlx.SqlConn) TC2cOrderModel {
	return &customTC2cOrderModel{
		defaultTC2cOrderModel: newTC2cOrderModel(conn),
	}
}

func (m *customTC2cOrderModel) withSession(session sqlx.Session) TC2cOrderModel {
	return NewTC2cOrderModel(sqlx.NewSqlConnFromSession(session))
}
