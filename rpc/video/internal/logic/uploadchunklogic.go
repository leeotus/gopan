// UploadChunkLogic 接收单个分片：存 MinIO + 标记 Redis 进度。
package logic

import (
	"bytes"
	"context"
	"fmt"

	"gopan/rpc/video/internal/svc"
	"gopan/rpc/video/video"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UploadChunkLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUploadChunkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadChunkLogic {
	return &UploadChunkLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *UploadChunkLogic) UploadChunk(in *video.UploadChunkReq) (*video.UploadChunkResp, error) {
	// 1. 上传分片到 MinIO
	key := fmt.Sprintf("parts/%d/chunk_%d", in.VideoId, in.ChunkIndex)
	err := l.svcCtx.MinioClient.PutObject(l.ctx, key, bytes.NewReader(in.Data), int64(in.FileSize), "application/octet-stream")
	if err != nil {
		l.Logger.Errorf("minio put chunk error: %v", err)
		return nil, status.Error(codes.Internal, "分片上传失败")
	}

	// 2. 标记 Redis 进度
	if err := l.svcCtx.UploadProgress.MarkReceived(l.ctx, in.UploadId, in.ChunkIndex); err != nil {
		l.Logger.Errorf("redis mark received error: %v", err)
		// 不阻断，upload-status 会暴露缺失
	}

	return &video.UploadChunkResp{ReceivedIndex: in.ChunkIndex}, nil
}
