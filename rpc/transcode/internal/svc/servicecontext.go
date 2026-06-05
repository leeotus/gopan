// package svc 定义 transcode-svc 的依赖注入容器。
package svc

import (
	"gopan/common/storage"
	"gopan/rpc/transcode/internal/config"
	"gopan/rpc/video/videoclient"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config      config.Config
	MinioClient *storage.MinioClient
	VideoClient videoclient.Video
	KafkaWriter *kafkago.Writer
}

func NewServiceContext(c config.Config) *ServiceContext {
	minioClient, _ := storage.NewMinioClient(
		c.MinIO.Endpoint, c.MinIO.AccessKey, c.MinIO.SecretKey,
		c.MinIO.Bucket, c.MinIO.UseSSL,
	)

	kw := &kafkago.Writer{
		Addr:     kafkago.TCP(c.Kafka.Brokers...),
		Topic:    c.Kafka.TranscodeTopic,
		Balancer: &kafkago.LeastBytes{},
	}

	return &ServiceContext{
		Config:      c,
		MinioClient: minioClient,
		VideoClient: videoclient.NewVideo(zrpc.MustNewClient(c.VideoRpc)),
		KafkaWriter: kw,
	}
}
