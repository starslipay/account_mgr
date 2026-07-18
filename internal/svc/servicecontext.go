package svc

import (
	"github.com/starslipay/account_mgr/internal/config"
	"github.com/starslipay/account_mgr/model/mysql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config               config.Config
	TCAccountModelMaster mysql.TCAccountModel
	TCAccountModelSlave  mysql.TCAccountModel

	TCAccountLogModelMaster mysql.TCAccountLogModel
	TCAccountLogModelSlave  mysql.TCAccountLogModel

	TC2crOrderMaster mysql.TC2cOrderModel
	TC2cOrderSlave   mysql.TC2cOrderModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	SqlMasterConn := sqlx.NewMysql(c.MasterDBConfig.DataSource)
	SqlSlaveConn := sqlx.NewMysql(c.SlaveDBConfig.DataSource)
	return &ServiceContext{
		Config:                  c,
		TCAccountModelMaster:    mysql.NewTCAccountModel(SqlMasterConn),
		TCAccountModelSlave:     mysql.NewTCAccountModel(SqlSlaveConn),
		TCAccountLogModelMaster: mysql.NewTCAccountLogModel(SqlMasterConn),
		TCAccountLogModelSlave:  mysql.NewTCAccountLogModel(SqlSlaveConn),
		TC2crOrderMaster:        mysql.NewTC2cOrderModel(SqlMasterConn),
		TC2cOrderSlave:          mysql.NewTC2cOrderModel(SqlSlaveConn),
	}
}
