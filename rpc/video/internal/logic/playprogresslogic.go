// PlayProgressLogic 保存/获取播放进度（Redis 存储）
package logic

import (
	"context"
	"fmt"
	"time"

	"gopan/rpc/video/internal/svc"
	"gopan/rpc/video/video"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SavePlayProgressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSavePlayProgressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SavePlayProgressLogic {
	return &SavePlayProgressLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *SavePlayProgressLogic) SavePlayProgress(in *video.SavePlayProgressReq) (*video.SavePlayProgressResp, error) {
	key := fmt.Sprintf("playback:%d:%d", in.UserId, in.VideoId)
	if err := l.svcCtx.PlaybackRedis.Set(l.ctx, key, in.Position, 30*24*time.Hour).Err(); err != nil {
		l.Logger.Errorf("redis save playback error: %v", err)
		return nil, status.Error(codes.Internal, "保存进度失败")
	}
	return &video.SavePlayProgressResp{}, nil
}

type GetPlayProgressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPlayProgressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPlayProgressLogic {
	return &GetPlayProgressLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetPlayProgressLogic) GetPlayProgress(in *video.GetPlayProgressReq) (*video.GetPlayProgressResp, error) {
	key := fmt.Sprintf("playback:%d:%d", in.UserId, in.VideoId)
	val, err := l.svcCtx.PlaybackRedis.Get(l.ctx, key).Float64()
	if err != nil {
		// key 不存在返回 0，不算错误
		if err.Error() == "redis: nil" {
			return &video.GetPlayProgressResp{Position: 0}, nil
		}
		return &video.GetPlayProgressResp{Position: 0}, nil
	}
	return &video.GetPlayProgressResp{Position: val}, nil
}
