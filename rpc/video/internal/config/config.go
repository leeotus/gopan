// package config 定义 video-svc（视频服务）的配置。
package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf               // RPC 通用配置
	SearchRpc  zrpc.RpcClientConf    // search-svc 客户端（ES 索引）
	DB         struct {              // MySQL 连接
		DataSource string
	}
	MinIO struct {                   // MinIO 对象存储
		Endpoint  string
		AccessKey string
		SecretKey string
		Bucket    string
		UseSSL    bool
	}
	Kafka struct {                   // Kafka 消息队列
		Brokers         []string
		TranscodeTopic  string
	}
	UploadRedis redis.RedisConf      // Redis 配置（上传进度追踪），避免与 go-zero 内置 Redis 冲突
}
