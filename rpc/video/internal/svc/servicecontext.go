// package svc 定义 video-svc 的依赖注入容器。
package svc

import (
	"gopan/common/kafka"
	"gopan/common/storage"
	"gopan/rpc/search/searchclient"
	"gopan/rpc/video/internal/config"
	"gopan/rpc/video/store"

	"github.com/redis/go-redis/v9"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	VideoStore     *store.VideoStore          // videos 和 transcodes 表的数据访问层
	MinioClient    *storage.MinioClient       // MinIO 对象存储客户端
	KafkaWriter      *kafkago.Writer            // Kafka Producer（转码任务）
	KafkaMergeWriter *kafkago.Writer            // Kafka Producer（合并任务）
	UploadProgress *storage.UploadProgress    // Redis 上传进度追踪
	PlaybackRedis  *redis.Client              // Redis 播放进度
	SearchClient   searchclient.Search        // search-svc gRPC 客户端
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

	kw := kafka.NewProducer(c.Kafka.Brokers, c.Kafka.TranscodeTopic)

		rdb := redis.NewClient(&redis.Options{
			Addr:     c.UploadRedis.Host,
			Password: c.UploadRedis.Pass,
			DB:       0,
		})

	return &ServiceContext{
		Config:         c,
		VideoStore:     store.NewVideoStore(conn),
		MinioClient:    minioClient,
		KafkaWriter:      kw,
		KafkaMergeWriter: kafka.NewProducer(c.Kafka.Brokers, c.Kafka.MergeTopic),
		UploadProgress: storage.NewUploadProgress(rdb),
		PlaybackRedis:  rdb,
		SearchClient:   searchclient.NewSearch(zrpc.MustNewClient(c.SearchRpc)),
	}
}
