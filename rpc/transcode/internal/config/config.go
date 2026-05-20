// package config 定义 transcode-svc（转码服务）的配置。
package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf        // RPC 通用配置
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
}
