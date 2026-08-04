package svc

import (
	"database/sql"
	"errors"
	"time"

	driverMysql "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/starslipay/account_mgr/internal/config"
	"github.com/starslipay/account_mgr/model/mysql"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// mysqlErrDuplicateEntry MySQL唯一键冲突错误码
const mysqlErrDuplicateEntry = 1062

// acceptDuplicateEntry 复刻 go-zero 的 mysqlAcceptable: 将1062唯一键冲突视为可接受错误,
// 使其不计入熔断失败(与原 sqlx.NewMysql 的熔断语义保持一致)
func acceptDuplicateEntry(err error) bool {
	if err == nil {
		return true
	}
	var mysqlErr *driverMysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlErrDuplicateEntry {
		return true
	}
	return false
}

// newDBConn 自建 *sql.DB 以支持自定义连接池参数, 并注册连接池 metrics,
// 最终用 NewSqlConnFromDB 包装(保留1062熔断豁免)
func newDBConn(dataSource, name string, maxOpen, maxIdle, lifetimeSec int) sqlx.SqlConn {
	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		logx.Must(err)
	}
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(time.Duration(lifetimeSec) * time.Second)
	if err = db.Ping(); err != nil {
		logx.Must(err)
	}
	prometheus.MustRegister(collectors.NewDBStatsCollector(db, name))
	return sqlx.NewSqlConnFromDB(db, sqlx.WithAcceptable(acceptDuplicateEntry))
}

type ServiceContext struct {
	Config               config.Config
	SqlMasterConn        sqlx.SqlConn
	SqlSlaveConn         sqlx.SqlConn
	TCAccountModelMaster mysql.TCAccountModel
	TCAccountModelSlave  mysql.TCAccountModel

	TCAccountLogModelMaster mysql.TCAccountLogModel
	TCAccountLogModelSlave  mysql.TCAccountLogModel

	TBAccountModelMaster mysql.TBAccountModel
	TBAccountModelSlave  mysql.TBAccountModel

	TBAccountLogModelMaster mysql.TBAccountLogModel
	TBAccountLogModelSlave  mysql.TBAccountLogModel

	TC2crOrderMaster mysql.TC2cOrderModel
	TC2cOrderSlave   mysql.TC2cOrderModel

	TSaveBillModelMaster mysql.TSaveBillModel
	TSaveBillModelSlave  mysql.TSaveBillModel

	TC2cBillModelMaster mysql.TC2cBillModel
	TC2cBillModelSlave  mysql.TC2cBillModel

	TC2cPendingTransferModelMaster mysql.TC2cPendingTransferModel
	TC2cPendingTransferModelSlave  mysql.TC2cPendingTransferModel

	TC2bBillModelMaster mysql.TC2bBillModel
	TC2bBillModelSlave  mysql.TC2bBillModel

	TC2bPendingTransferModelMaster mysql.TC2bPendingTransferModel
	TC2bPendingTransferModelSlave  mysql.TC2bPendingTransferModel

	C2CAsyncTransferProducer *kq.Pusher
	C2BAsyncTransferProducer *kq.Pusher
}

func NewServiceContext(c config.Config) *ServiceContext {
	SqlMasterConn := newDBConn(c.MasterDBConfig.DataSource, "account_master",
		c.MasterDBConfig.MaxOpenConns, c.MasterDBConfig.MaxIdleConns, c.MasterDBConfig.ConnMaxLifetimeSec)
	SqlSlaveConn := newDBConn(c.SlaveDBConfig.DataSource, "account_slave",
		c.SlaveDBConfig.MaxOpenConns, c.SlaveDBConfig.MaxIdleConns, c.SlaveDBConfig.ConnMaxLifetimeSec)

	return &ServiceContext{
		Config:                         c,
		SqlMasterConn:                  SqlMasterConn,
		SqlSlaveConn:                   SqlSlaveConn,
		TCAccountModelMaster:           mysql.NewTCAccountModel(SqlMasterConn),
		TCAccountModelSlave:            mysql.NewTCAccountModel(SqlSlaveConn),
		TCAccountLogModelMaster:        mysql.NewTCAccountLogModel(SqlMasterConn),
		TCAccountLogModelSlave:         mysql.NewTCAccountLogModel(SqlSlaveConn),
		TBAccountModelMaster:           mysql.NewTBAccountModel(SqlMasterConn),
		TBAccountModelSlave:            mysql.NewTBAccountModel(SqlSlaveConn),
		TBAccountLogModelMaster:        mysql.NewTBAccountLogModel(SqlMasterConn),
		TBAccountLogModelSlave:         mysql.NewTBAccountLogModel(SqlSlaveConn),
		TC2crOrderMaster:               mysql.NewTC2cOrderModel(SqlMasterConn),
		TC2cOrderSlave:                 mysql.NewTC2cOrderModel(SqlSlaveConn),
		TSaveBillModelMaster:           mysql.NewTSaveBillModel(SqlMasterConn),
		TSaveBillModelSlave:            mysql.NewTSaveBillModel(SqlSlaveConn),
		TC2cBillModelMaster:            mysql.NewTC2cBillModel(SqlMasterConn),
		TC2cBillModelSlave:             mysql.NewTC2cBillModel(SqlSlaveConn),
		TC2cPendingTransferModelMaster: mysql.NewTC2cPendingTransferModel(SqlMasterConn),
		TC2cPendingTransferModelSlave:  mysql.NewTC2cPendingTransferModel(SqlSlaveConn),
		TC2bBillModelMaster:            mysql.NewTC2bBillModel(SqlMasterConn),
		TC2bBillModelSlave:             mysql.NewTC2bBillModel(SqlSlaveConn),
		TC2bPendingTransferModelMaster: mysql.NewTC2bPendingTransferModel(SqlMasterConn),
		TC2bPendingTransferModelSlave:  mysql.NewTC2bPendingTransferModel(SqlSlaveConn),
		C2CAsyncTransferProducer:       kq.NewPusher(c.KafkaProducerConf, c.TopicC2cAsyncTransfer),
		C2BAsyncTransferProducer:       kq.NewPusher(c.KafkaProducerConf, c.TopicC2bAsyncTransfer),
	}
}
