// ListVideosLogic 视频列表（游标分页，支持分类和排序）。
// cursor=0 取第一页，has_more 通过多查一条判断。
package logic

import (
	"context"

	"gopan/rpc/video/internal/svc"
	"gopan/rpc/video/video"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ListVideosLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListVideosLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVideosLogic {
	return &ListVideosLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListVideosLogic) ListVideos(in *video.ListVideosReq) (*video.ListVideosResp, error) {
	videos, err := l.svcCtx.VideoStore.List(l.ctx, in.Cursor, in.Limit, in.Category, in.Sort)
	if err != nil {
		l.Logger.Errorf("list videos error: %v", err)
		return nil, status.Error(codes.Internal, "查询失败")
	}

	hasMore := len(videos) > int(in.Limit)
	if hasMore {
		videos = videos[:in.Limit] // 去掉多查的那一条
	}

	resp := &video.ListVideosResp{HasMore: hasMore}
	for _, v := range videos {
		info := toVideoInfo(v)
		transcodes, _ := l.svcCtx.VideoStore.GetTranscodes(l.ctx, v.Id)
		for _, t := range transcodes {
			info.Transcodes = append(info.Transcodes, &video.TranscodeInfo{
				Resolution: t.Resolution,
				M3U8Url:    t.M3U8Url,
				Bitrate:    t.Bitrate,
			})
		}
		resp.Videos = append(resp.Videos, info)
		if !hasMore {
			resp.NextCursor = v.Id // 最后一条的 ID 作为下一页游标
		}
	}
	return resp, nil
}
