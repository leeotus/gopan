// MergeChunksLogic 合并上传分片（暂为桩实现，需集成 MinIO 后才完整）。
package logic

import (
	"context"

	"gopan/rpc/video/internal/svc"
	"gopan/rpc/video/video"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MergeChunksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMergeChunksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MergeChunksLogic {
	return &MergeChunksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MergeChunksLogic) MergeChunks(in *video.MergeChunksReq) (*video.MergeChunksResp, error) {
	// 合并完成后更新视频状态为"转码中"
	if err := l.svcCtx.VideoStore.UpdateStatus(l.ctx, in.VideoId, 1); err != nil {
		return nil, status.Error(codes.Internal, "更新视频状态失败")
	}

	v, err := l.svcCtx.VideoStore.FindById(l.ctx, in.VideoId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "视频不存在")
	}

	return &video.MergeChunksResp{
		ObjectKey: v.ObjectKey,
		FileHash:  v.FileHash,
	}, nil
}
