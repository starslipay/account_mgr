package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	MasterDBConfig struct {
		DataSource string
	}
	SlaveDBConfig struct {
		DataSource string
	}
	KafkaProducerConf     []string
	TopicC2cAsyncTransfer string
	TopicC2bAsyncTransfer string
}
