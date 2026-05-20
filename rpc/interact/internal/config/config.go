// package config 定义 interact-svc（互动服务）的配置。
package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf       // RPC 通用配置
	DB          struct {     // MySQL 连接
		DataSource string
	}
	CacheRedis  struct {     // Redis 缓存（点赞数、弹幕广播等）
		Host     string
		Port     int
		Password string
		DB       int
	}
}
