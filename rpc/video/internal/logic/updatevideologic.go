// UpdateVideoLogic 更新视频标题、简介、分类。
// 仅允许视频所有者操作（由调用方保证权限）。
package logic

import (
	"context"

	"gopan/rpc/video/internal/svc"
	"gopan/rpc/video/model"
	"gopan/rpc/video/video"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpdateVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateVideoLogic {
	return &UpdateVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateVideoLogic) UpdateVideo(in *video.UpdateVideoReq) (*video.UpdateVideoResp, error) {
	v := &model.Video{
		Id:          in.VideoId,
		Title:       in.Title,
		Description: in.Description,
		Category:    in.Category,
	}
	if err := l.svcCtx.VideoStore.Update(l.ctx, v); err != nil {
		l.Logger.Errorf("update video error: %v", err)
		return nil, status.Error(codes.Internal, "更新失败")
	}
	return &video.UpdateVideoResp{}, nil
}
