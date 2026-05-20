// package config 定义 video-svc（视频服务）的配置。
package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf          // RPC 通用配置
	DB    struct {              // MySQL 连接
		DataSource string
	}
	MinIO struct {              // MinIO 对象存储
		Endpoint  string         // 地址，如 minio:9000
		AccessKey string
		SecretKey string
		Bucket    string         // 存储桶名称
		UseSSL    bool
	}
}
