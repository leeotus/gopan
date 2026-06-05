// UpdateVideoLogic 更新视频标题、简介、分类。
// 仅允许视频所有者操作（由调用方保证权限）。
package logic

import (
	"context"

	"gopan/rpc/video/internal/svc"
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
	// 1. 先从数据库查询出完整、原装的视频旧记录
	v, err := l.svcCtx.VideoStore.FindById(l.ctx, in.VideoId)
	if err != nil {
		l.Logger.Errorf("find video %d error: %v", in.VideoId, err)
		return nil, status.Error(codes.NotFound, "视频不存在")
	}

	// 2. 仅对入参中非空的字段进行局部刷新，绝不采用空值覆盖、抹除 CoverUrl, PlayCount 和 Status 状态等黄金原始数据列！
	if in.Title != "" {
		v.Title = in.Title
	}
	if in.Description != "" {
		v.Description = in.Description
	}
	if in.Category != "" {
		v.Category = in.Category
	}

	// 3. 将包含原装数据且被局部更新后的完整对象安全写入 DB
	if err := l.svcCtx.VideoStore.Update(l.ctx, v); err != nil {
		l.Logger.Errorf("update video error: %v", err)
		return nil, status.Error(codes.Internal, "更新失败")
	}
	return &video.UpdateVideoResp{}, nil
}
