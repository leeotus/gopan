// DeleteVideoLogic 软删除视频（设置 deleted_at），不物理删除数据。
package logic

import (
	"context"

	"gopan/rpc/video/internal/svc"
	"gopan/rpc/video/video"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DeleteVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteVideoLogic {
	return &DeleteVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteVideoLogic) DeleteVideo(in *video.DeleteVideoReq) (*video.DeleteVideoResp, error) {
	if err := l.svcCtx.VideoStore.Delete(l.ctx, in.VideoId, in.UserId); err != nil {
		l.Logger.Errorf("delete video error: %v", err)
		return nil, status.Error(codes.Internal, "删除失败")
	}
	return &video.DeleteVideoResp{}, nil
}
