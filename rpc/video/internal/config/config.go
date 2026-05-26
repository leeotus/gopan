package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	SearchRpc   zrpc.RpcClientConf
	DB          struct { DataSource string }
	MinIO       struct { Endpoint, AccessKey, SecretKey, Bucket string; UseSSL bool }
	Kafka       struct { Brokers []string; TranscodeTopic string }
	UploadRedis struct { Host, Pass string }
}
