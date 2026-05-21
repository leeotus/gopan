// package storage 提供 MinIO 对象存储的通用封装，video-svc 和 transcode-svc 共享。
package storage

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioClient 封装 MinIO 客户端和 Bucket 信息。
type MinioClient struct {
	client *minio.Client
	bucket string
}

// NewMinioClient 创建 MinIO 客户端并确保 Bucket 存在。
func NewMinioClient(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinioClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio connect failed: %w", err)
	}

	exists, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		return nil, fmt.Errorf("minio bucket check failed: %w", err)
	}
	if !exists {
		err = client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("minio create bucket failed: %w", err)
		}
		log.Printf("[minio] bucket '%s' created", bucket)
	}

	return &MinioClient{client: client, bucket: bucket}, nil
}

// PutObject 上传文件到 MinIO。
// 返回 key 的完整路径（供后续转码任务等使用）。
func (s *MinioClient) PutObject(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// GetObject 从 MinIO 下载文件。
func (s *MinioClient) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	return s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
}

// ObjectURL 生成外部访问 URL。
func (s *MinioClient) ObjectURL(key string) string {
	return fmt.Sprintf("http://%s/%s/%s", s.client.EndpointURL().Host, s.bucket, key)
}

// RemoveObject 删除对象。
func (s *MinioClient) RemoveObject(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

// BucketName 返回当前使用的 bucket 名称。
func (s *MinioClient) BucketName() string {
	return s.bucket
}
