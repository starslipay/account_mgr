package svc

import (
	"github.com/starslipay/account_mgr/internal/config"
	"github.com/starslipay/account_mgr/model/mysql"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config               config.Config
	SqlMasterConn        sqlx.SqlConn
	SqlSlaveConn         sqlx.SqlConn
	TCAccountModelMaster mysql.TCAccountModel
	TCAccountModelSlave  mysql.TCAccountModel

	TCAccountLogModelMaster mysql.TCAccountLogModel
	TCAccountLogModelSlave  mysql.TCAccountLogModel

	TC2crOrderMaster mysql.TC2cOrderModel
	TC2cOrderSlave   mysql.TC2cOrderModel

	TSaveBillModelMaster mysql.TSaveBillModel
	TSaveBillModelSlave  mysql.TSaveBillModel

	TC2cBillModelMaster mysql.TC2cBillModel
	TC2cBillModelSlave  mysql.TC2cBillModel

	TC2cPendingTransferModelMaster mysql.TC2cPendingTransferModel
	TC2cPendingTransferModelSlave  mysql.TC2cPendingTransferModel

	C2CAsyncTransferProducer *kq.Pusher
}

func NewServiceContext(c config.Config) *ServiceContext {
	SqlMasterConn := sqlx.NewMysql(c.MasterDBConfig.DataSource)
	SqlSlaveConn := sqlx.NewMysql(c.SlaveDBConfig.DataSource)

	return &ServiceContext{
		Config:                         c,
		SqlMasterConn:                  SqlMasterConn,
		SqlSlaveConn:                   SqlSlaveConn,
		TCAccountModelMaster:           mysql.NewTCAccountModel(SqlMasterConn),
		TCAccountModelSlave:            mysql.NewTCAccountModel(SqlSlaveConn),
		TCAccountLogModelMaster:        mysql.NewTCAccountLogModel(SqlMasterConn),
		TCAccountLogModelSlave:         mysql.NewTCAccountLogModel(SqlSlaveConn),
		TC2crOrderMaster:               mysql.NewTC2cOrderModel(SqlMasterConn),
		TC2cOrderSlave:                 mysql.NewTC2cOrderModel(SqlSlaveConn),
		TSaveBillModelMaster:           mysql.NewTSaveBillModel(SqlMasterConn),
		TSaveBillModelSlave:            mysql.NewTSaveBillModel(SqlSlaveConn),
		TC2cBillModelMaster:            mysql.NewTC2cBillModel(SqlMasterConn),
		TC2cBillModelSlave:             mysql.NewTC2cBillModel(SqlSlaveConn),
		TC2cPendingTransferModelMaster: mysql.NewTC2cPendingTransferModel(SqlMasterConn),
		TC2cPendingTransferModelSlave:  mysql.NewTC2cPendingTransferModel(SqlSlaveConn),
		C2CAsyncTransferProducer:       kq.NewPusher(c.KafkaProducerConf, c.TopicC2cAsyncTransfer),
	}
}
