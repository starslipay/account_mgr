package mysql

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ TPendingC2cTransferModel = (*customTPendingC2cTransferModel)(nil)

type (
	// TPendingC2cTransferModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTPendingC2cTransferModel.
	TPendingC2cTransferModel interface {
		tPendingC2cTransferModel
		withSession(session sqlx.Session) TPendingC2cTransferModel
	}

	customTPendingC2cTransferModel struct {
		*defaultTPendingC2cTransferModel
	}
)

// NewTPendingC2cTransferModel returns a model for the database table.
func NewTPendingC2cTransferModel(conn sqlx.SqlConn) TPendingC2cTransferModel {
	return &customTPendingC2cTransferModel{
		defaultTPendingC2cTransferModel: newTPendingC2cTransferModel(conn),
	}
}

func (m *customTPendingC2cTransferModel) withSession(session sqlx.Session) TPendingC2cTransferModel {
	return NewTPendingC2cTransferModel(sqlx.NewSqlConnFromSession(session))
}
