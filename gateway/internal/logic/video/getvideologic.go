// GetVideoLogic 视频详情编排。
package video

import (
	"context"
	"strconv"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/video/videoclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVideoLogic {
	return &GetVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetVideoLogic) GetVideo(videoIdStr string) (resp *types.VideoDetailResp, err error) {
	videoId, _ := strconv.ParseInt(videoIdStr, 10, 64)
	r, err := l.svcCtx.VideoClient.GetVideo(l.ctx, &videoclient.GetVideoReq{VideoId: videoId})
	if err != nil {
		return nil, err
	}

	v := r.Video
	transcodes := make([]types.TranscodeInfo, 0, len(v.Transcodes))
	for _, t := range v.Transcodes {
		transcodes = append(transcodes, types.TranscodeInfo{
			Resolution: t.Resolution,
			M3u8Url:    t.M3U8Url,
			Bitrate:    int(t.Bitrate),
		})
	}
	return &types.VideoDetailResp{
		Video: types.VideoInfo{
			Id:          v.Id,
			Title:       v.Title,
			CoverUrl:    v.CoverUrl,
			UserId:      v.UserId,
			PlayCount:   v.PlayCount,
			LikeCount:   v.LikeCount,
			Duration:    int(v.Duration),
			Status:      int(v.Status),
			Category:    v.Category,
			Description: v.Description,
			CreatedAt:   v.CreatedAt,
			Transcodes:  transcodes,
		},
	}, nil
}
