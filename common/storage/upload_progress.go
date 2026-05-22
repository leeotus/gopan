// package storage 提供上传进度的 Redis 操作（分片断点上传用）。
package storage

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// UploadProgress 管理分片上传进度（Redis Set）。
type UploadProgress struct {
	rdb *redis.Client
}

// NewUploadProgress 创建进度管理器。
func NewUploadProgress(rdb *redis.Client) *UploadProgress {
	return &UploadProgress{rdb: rdb}
}

// MarkReceived 标记 chunk_index 已收到。
func (p *UploadProgress) MarkReceived(ctx context.Context, uploadId string, index int32) error {
	return p.rdb.SAdd(ctx, progressKey(uploadId), index).Err()
}

// GetReceived 获取所有已收到的 chunk index 列表。
func (p *UploadProgress) GetReceived(ctx context.Context, uploadId string) ([]int32, error) {
	vals, err := p.rdb.SMembers(ctx, progressKey(uploadId)).Result()
	if err != nil {
		return nil, err
	}
	var indexes []int32
	for _, v := range vals {
		var i int
		fmt.Sscanf(v, "%d", &i)
		indexes = append(indexes, int32(i))
	}
	return indexes, nil
}

// CountReceived 返回已收到的 chunk 数量。
func (p *UploadProgress) CountReceived(ctx context.Context, uploadId string) (int64, error) {
	return p.rdb.SCard(ctx, progressKey(uploadId)).Result()
}

// Clear 清理进度数据（合并完成后调用）。
func (p *UploadProgress) Clear(ctx context.Context, uploadId string) error {
	return p.rdb.Del(ctx, progressKey(uploadId)).Err()
}

func progressKey(uploadId string) string {
	return fmt.Sprintf("upload:%s:received", uploadId)
}
