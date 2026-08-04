package config

import "github.com/zeromicro/go-zero/zrpc"

type DBConfig struct {
	DataSource         string // 数据库连接字符串
	MaxOpenConns       int    // 最大打开连接数
	MaxIdleConns       int    // 最大空闲连接数
	ConnMaxLifetimeSec int    // 连接最大生命周期秒数
}

type Config struct {
	zrpc.RpcServerConf
	MasterDBConfig        DBConfig
	SlaveDBConfig         DBConfig
	KafkaProducerConf     []string
	TopicC2cAsyncTransfer string
	TopicC2bAsyncTransfer string
}
