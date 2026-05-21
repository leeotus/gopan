// package svc 定义 video-svc 的依赖注入容器。
package svc

import (
	"gopan/common/storage"
	"gopan/rpc/search/searchclient"
	"gopan/rpc/video/internal/config"
	"gopan/rpc/video/store"

	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config       config.Config
	VideoStore   *store.VideoStore       // videos 和 transcodes 表的共享数据访问层
	MinioClient  *storage.MinioClient    // MinIO 对象存储客户端
	KafkaWriter  *kafka.Writer           // Kafka Producer，发送转码任务消息
	SearchClient searchclient.Search     // search-svc gRPC 客户端（ES 索引）
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DB.DataSource)

	minioClient, err := storage.NewMinioClient(
		c.MinIO.Endpoint, c.MinIO.AccessKey, c.MinIO.SecretKey,
		c.MinIO.Bucket, c.MinIO.UseSSL,
	)
	if err != nil {
		logx.Errorf("minio init failed: %v", err)
	}

	kw := &kafka.Writer{
		Addr:     kafka.TCP(c.Kafka.Brokers...),
		Topic:    c.Kafka.TranscodeTopic,
		Balancer: &kafka.LeastBytes{},
		Logger:   nil,
	}

	return &ServiceContext{
		Config:       c,
		VideoStore:   store.NewVideoStore(conn),
		MinioClient:  minioClient,
		KafkaWriter:  kw,
		SearchClient: searchclient.NewSearch(zrpc.MustNewClient(c.SearchRpc)),
	}
}
