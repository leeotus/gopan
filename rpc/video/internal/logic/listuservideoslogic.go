// ListUserVideosLogic 获取指定用户上传的视频列表。
package logic

import (
	"context"

	"gopan/rpc/video/internal/svc"
	"gopan/rpc/video/video"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ListUserVideosLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserVideosLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserVideosLogic {
	return &ListUserVideosLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListUserVideosLogic) ListUserVideos(in *video.ListUserVideosReq) (*video.ListVideosResp, error) {
	videos, err := l.svcCtx.VideoStore.ListByUser(l.ctx, in.UserId, in.Cursor, in.Limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "查询失败")
	}

	hasMore := len(videos) > int(in.Limit)
	if hasMore {
		videos = videos[:in.Limit]
	}

	resp := &video.ListVideosResp{HasMore: hasMore}
	for _, v := range videos {
		resp.Videos = append(resp.Videos, toVideoInfo(v))
		if !hasMore {
			resp.NextCursor = v.Id
		}
	}
	return resp, nil
}
