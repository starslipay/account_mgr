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
	Kafka struct {
		BrokerAddrs []string
		Topic       string
	}
}
