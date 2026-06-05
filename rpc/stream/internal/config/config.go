package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	CDN        struct{ BaseURL, SecretKey string }
	CacheRedis struct {
		Host     string
		Port     int
		Password string
		DB       int
	}
}
