// ListVideosLogic 视频列表（游标分页，支持分类和排序）。
// cursor=0 取第一页，has_more 通过多查一条判断。
package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
	// 首页第一页 (cursor=0, sort=newest) 走 Redis 缓存，60 秒过期
	if in.Cursor == 0 && in.Category == "" && in.Sort == "newest" {
		cacheKey := fmt.Sprintf("video:list:home:limit:%d", in.Limit)
		if cached := l.getCached(cacheKey); cached != nil {
			l.Logger.Infof("list videos from cache, key=%s", cacheKey)
			return cached, nil
		}
		resp, err := l.fetchFromDB(in)
		if err == nil && resp != nil {
			l.setCached(cacheKey, resp, 60*time.Second)
		}
		l.Logger.Infof("list videos from db, key=%s", cacheKey)
		return resp, err
	}

	return l.fetchFromDB(in)
}

func (l *ListVideosLogic) fetchFromDB(in *video.ListVideosReq) (*video.ListVideosResp, error) {
	videos, err := l.svcCtx.VideoStore.List(l.ctx, in.Cursor, in.Limit, in.Category, in.Sort)
	if err != nil {
		l.Logger.Errorf("list videos error: %v", err)
		return nil, status.Error(codes.Internal, "查询失败")
	}

	hasMore := len(videos) > int(in.Limit)
	if hasMore {
		videos = videos[:in.Limit]
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
			resp.NextCursor = v.Id
		}
	}
	return resp, nil
}

func (l *ListVideosLogic) getCached(key string) *video.ListVideosResp {
	data, err := l.svcCtx.PlaybackRedis.Get(l.ctx, key).Bytes()
	if err != nil {
		return nil
	}
	var resp video.ListVideosResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil
	}
	return &resp
}

func (l *ListVideosLogic) setCached(key string, resp *video.ListVideosResp, ttl time.Duration) {
	data, _ := json.Marshal(resp)
	_ = l.svcCtx.PlaybackRedis.Set(l.ctx, key, data, ttl).Err()
}
