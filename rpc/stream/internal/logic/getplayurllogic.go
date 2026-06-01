// GetPlayUrlLogic 生成播放地址（带防盗链签名）。
// 签名方案: MD5(video_id + user_id + expire + secret) → 时间戳过期校验。
package logic

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	"gopan/rpc/stream/internal/svc"
	"gopan/rpc/stream/stream"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPlayUrlLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPlayUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPlayUrlLogic {
	return &GetPlayUrlLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetPlayUrl 生成带防盗链签名的播放地址。
// 返回多码率 HLS 地址列表，前端播放器可自动切换清晰度。
func (l *GetPlayUrlLogic) GetPlayUrl(in *stream.GetPlayUrlReq) (*stream.GetPlayUrlResp, error) {
	baseURL := l.svcCtx.Config.CDN.BaseURL
	secretKey := l.svcCtx.Config.CDN.SecretKey

	// 生成 MD5 签名: sign = md5(videoId-userId-expire-secretKey)
	expire := time.Now().Add(2 * time.Hour).Unix()
	signStr := fmt.Sprintf("%d-%d-%d-%s", in.VideoId, in.UserId, expire, secretKey)
	sign := fmt.Sprintf("%x", md5.Sum([]byte(signStr)))

	resolutions := []struct {
		name    string
		bitrate int32
	}{
		{"360p", 500},
		{"480p", 1000},
		{"720p", 2500},
		{"1080p", 5000},
	}

	resp := &stream.GetPlayUrlResp{
		M3U8Url: fmt.Sprintf("%s/videos/%d/playlist.m3u8?sign=%s&expire=%d",
			baseURL, in.VideoId, sign, expire),
	}

	for _, r := range resolutions {
		if in.Resolution != "" && in.Resolution != r.name {
			continue // 指定了分辨率则只返回该档
		}
		resp.Streams = append(resp.Streams, &stream.PlayStream{
			Resolution: r.name,
			Url: fmt.Sprintf("%s/videos/%d/%s/index.m3u8?sign=%s&expire=%d",
				baseURL, in.VideoId, r.name, sign, expire),
			Bitrate: r.bitrate,
		})
	}

	return resp, nil
}

// IncrPlayCountLogic 增加播放计数（Redis INCR + 每 100 次同步 MySQL）。
type IncrPlayCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIncrPlayCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IncrPlayCountLogic {
	return &IncrPlayCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IncrPlayCountLogic) IncrPlayCount(in *stream.IncrPlayCountReq) (*stream.IncrPlayCountResp, error) {
	key := fmt.Sprintf("video:play:%d", in.VideoId)
	count, err := l.svcCtx.Redis.Incr(l.ctx, key).Result()
	if err != nil {
		l.Logger.Errorf("redis incr play count error: %v", err)
		return &stream.IncrPlayCountResp{PlayCount: 0}, nil
	}
	return &stream.IncrPlayCountResp{PlayCount: count}, nil
}
