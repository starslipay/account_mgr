package svc

import (
	"github.com/IBM/sarama"
	"github.com/starslipay/account_mgr/internal/config"
	"github.com/starslipay/account_mgr/model/mysql"
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

	TLocalMessageModelMaster mysql.TLocalMessageModel
	TLocalMessageModelSlave  mysql.TLocalMessageModel

	KafkaProducer sarama.SyncProducer
}

func NewServiceContext(c config.Config) *ServiceContext {
	SqlMasterConn := sqlx.NewMysql(c.MasterDBConfig.DataSource)
	SqlSlaveConn := sqlx.NewMysql(c.SlaveDBConfig.DataSource)

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.Return.Errors = true
	kafkaConfig.Version = sarama.V3_8_0_0

	kafkaProducer, err := sarama.NewSyncProducer(c.Kafka.BrokerAddrs, kafkaConfig)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:                   c,
		SqlMasterConn:            SqlMasterConn,
		SqlSlaveConn:             SqlSlaveConn,
		TCAccountModelMaster:     mysql.NewTCAccountModel(SqlMasterConn),
		TCAccountModelSlave:      mysql.NewTCAccountModel(SqlSlaveConn),
		TCAccountLogModelMaster:  mysql.NewTCAccountLogModel(SqlMasterConn),
		TCAccountLogModelSlave:   mysql.NewTCAccountLogModel(SqlSlaveConn),
		TC2crOrderMaster:         mysql.NewTC2cOrderModel(SqlMasterConn),
		TC2cOrderSlave:           mysql.NewTC2cOrderModel(SqlSlaveConn),
		TSaveBillModelMaster:     mysql.NewTSaveBillModel(SqlMasterConn),
		TSaveBillModelSlave:      mysql.NewTSaveBillModel(SqlSlaveConn),
		TC2cBillModelMaster:      mysql.NewTC2cBillModel(SqlMasterConn),
		TC2cBillModelSlave:       mysql.NewTC2cBillModel(SqlSlaveConn),
		TLocalMessageModelMaster: mysql.NewTLocalMessageModel(SqlMasterConn),
		TLocalMessageModelSlave:  mysql.NewTLocalMessageModel(SqlSlaveConn),
		KafkaProducer:            kafkaProducer,
	}
}
