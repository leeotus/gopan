// package config 定义 transcode-svc（转码服务）的配置。
package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf        // RPC 通用配置
	VideoRpc zrpc.RpcClientConf // video-svc 的客户端配置（回调用）

	FFmpeg struct {           // FFmpeg 可执行文件路径
		Path string            // 如 /usr/bin/ffmpeg
	}
	MinIO struct {            // MinIO 对象存储（读取源文件、写入 HLS 切片）
		Endpoint  string
		AccessKey string
		SecretKey string
		Bucket    string
		UseSSL    bool
	}
	Kafka struct {                    // Kafka 消息队列
		Brokers        []string        // 如 ["kafka:9092"]
		TranscodeTopic string          // 消费 topic "gopan.transcode.tasks"
		ConsumerGroup  string          // 消费者组 "transcode-svc-group"
	}
	// 转码输出目录的临时工作路径
	WorkDir string `json:",default=/tmp/gopan-transcode"`
}
