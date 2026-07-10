package mysql

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ TC2cTransferOrderModel = (*customTC2cTransferOrderModel)(nil)

type (
	// TC2cTransferOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTC2cTransferOrderModel.
	TC2cTransferOrderModel interface {
		tC2cTransferOrderModel
		withSession(session sqlx.Session) TC2cTransferOrderModel
	}

	customTC2cTransferOrderModel struct {
		*defaultTC2cTransferOrderModel
	}
)

// NewTC2cTransferOrderModel returns a model for the database table.
func NewTC2cTransferOrderModel(conn sqlx.SqlConn) TC2cTransferOrderModel {
	return &customTC2cTransferOrderModel{
		defaultTC2cTransferOrderModel: newTC2cTransferOrderModel(conn),
	}
}

func (m *customTC2cTransferOrderModel) withSession(session sqlx.Session) TC2cTransferOrderModel {
	return NewTC2cTransferOrderModel(sqlx.NewSqlConnFromSession(session))
}
