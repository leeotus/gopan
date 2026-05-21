// package svc 定义 transcode-svc 的依赖注入容器。
package svc

import (
	"gopan/common/storage"
	"gopan/rpc/transcode/internal/config"
	"gopan/rpc/video/videoclient"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config      config.Config
	MinioClient *storage.MinioClient // MinIO 客户端（下载源文件 + 上传 HLS 切片）
	VideoClient videoclient.Video    // video-svc gRPC 客户端（回调用）
}

func NewServiceContext(c config.Config) *ServiceContext {
	minioClient, err := storage.NewMinioClient(
		c.MinIO.Endpoint, c.MinIO.AccessKey, c.MinIO.SecretKey,
		c.MinIO.Bucket, c.MinIO.UseSSL,
	)
	if err != nil {
		logx.Errorf("minio init failed: %v", err)
	}

	return &ServiceContext{
		Config:      c,
		MinioClient: minioClient,
		VideoClient: videoclient.NewVideo(zrpc.MustNewClient(c.VideoRpc)),
	}
}
