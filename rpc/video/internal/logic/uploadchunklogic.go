// UploadChunkLogic 分片上传逻辑（暂为桩实现）。
// 完整实现需：接收分片 → 写临时目录 → 全部到达后合并 → 上传到 MinIO。
package logic

import (
	"context"

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
	return &UploadChunkLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UploadChunk 处理客户端流式分片上传。
// 当前为桩，返回 Unimplemented。
func (l *UploadChunkLogic) UploadChunk(stream video.Video_UploadChunkServer) error {
	return status.Error(codes.Unimplemented, "分片上传请使用 HTTP 接口 /api/video/init-upload + /api/video/upload-chunk")
}
