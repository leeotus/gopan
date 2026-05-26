package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	DB         struct { DataSource string }
	CacheRedis struct { Host string; Port int; Password string; DB int }
}
