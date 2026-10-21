package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

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
	TopicUpdateBalance    string
	UpdateBalanceCmq      UpdateBalanceCmqConf
	RedisConfig           redis.RedisConf

	// AccessLog 请求/响应日志脱敏配置
	AccessLog AccessLogConf
}

// UpdateBalanceCmqConf 余额缓存更新消息开关配置
type UpdateBalanceCmqConf struct {
	Enable bool `json:",default=true"` // 是否发送 cmq(t_update_balance) 更新余额缓存
}

// AccessLogConf 访问日志配置
type AccessLogConf struct {
	Enable          bool     `json:",default=true"`
	SensitiveFields []string `json:",optional"`
}
