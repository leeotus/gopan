// ListVideosLogic 视频列表编排。
package video

import (
	"context"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/video/videoclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListVideosLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListVideosLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVideosLogic {
	return &ListVideosLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListVideosLogic) ListVideos(req *types.ListVideosReq) (resp *types.ListVideosResp, err error) {
	r, err := l.svcCtx.VideoClient.ListVideos(l.ctx, &videoclient.ListVideosReq{
		Cursor:   req.Cursor,
		Limit:    int32(req.Limit),
		Category: req.Category,
		Sort:     req.Sort,
	})
	if err != nil {
		return nil, err
	}

	videos := make([]types.VideoInfo, 0, len(r.Videos))
	for _, v := range r.Videos {
		transcodes := make([]types.TranscodeInfo, 0, len(v.Transcodes))
		for _, t := range v.Transcodes {
			transcodes = append(transcodes, types.TranscodeInfo{
				Resolution: t.Resolution,
				M3u8Url:    t.M3U8Url,
				Bitrate:    int(t.Bitrate),
			})
		}
			videos = append(videos, types.VideoInfo{
				Id:              v.Id,
				Title:           v.Title,
				CoverUrl:        v.CoverUrl,
				UserId:          v.UserId,
				PlayCount:       v.PlayCount,
				LikeCount:       v.LikeCount,
				Duration:        int(v.Duration),
				Status:          int(v.Status),
				Category:        v.Category,
				Description:     v.Description,
				AiSummary:       v.AiSummary,
				AiSummaryStatus: int(v.AiSummaryStatus),
				CreatedAt:       v.CreatedAt,
				Transcodes:      transcodes,
			})
	}
	return &types.ListVideosResp{
		Videos:     videos,
		NextCursor: r.NextCursor,
		HasMore:    r.HasMore,
	}, nil
}
