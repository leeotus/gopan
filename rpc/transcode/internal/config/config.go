package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	VideoRpc  zrpc.RpcClientConf
	FFmpeg    struct { Path string }
	MinIO     struct { Endpoint, AccessKey, SecretKey, Bucket string; UseSSL bool }
	Kafka     struct { Brokers []string; TranscodeTopic string }
	WorkDir   string `json:",default=/tmp/gopan-transcode"`
}
