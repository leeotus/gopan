package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	SearchRpc   zrpc.RpcClientConf
	DB          struct{ DataSource string }
	MinIO       struct {
		Endpoint, AccessKey, SecretKey, Bucket string
		UseSSL                                 bool
	}
	Kafka struct {
		Brokers        []string
		TranscodeTopic string
		MergeTopic     string
		SummaryTopic   string // AI 摘要任务 topic
	}
	UploadRedis struct{ Host, Pass string }
	SummaryAI   struct {
		URL     string // summary-ai 服务地址，例如 http://127.0.0.1:9920
		MinIO   string // 视频文件可下载的 MinIO 前缀，例如 http://127.0.0.1:9000/gopan-videos
		Timeout int    // /analyze 调用超时秒数，默认 300
	}
}
