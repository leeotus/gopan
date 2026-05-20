// SendDanmakuLogic / GetDanmakusLogic — 弹幕相关逻辑。
// TODO: 集成 danmakus 表 + Redis Pub/Sub + WebSocket 实时推送。
package logic

import (
	"context"

	"gopan/rpc/interact/internal/svc"
	"gopan/rpc/interact/interact"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SendDanmakuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendDanmakuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendDanmakuLogic {
	return &SendDanmakuLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *SendDanmakuLogic) SendDanmaku(in *interact.SendDanmakuReq) (*interact.SendDanmakuResp, error) {
	l.Logger.Infof("send danmaku: videoId=%d, content=%s, time=%.1f", in.VideoId, in.Content, in.Time)
	return &interact.SendDanmakuResp{}, status.Error(codes.Unimplemented, "需要 Redis Pub/Sub + WebSocket")
}

type GetDanmakusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDanmakusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDanmakusLogic {
	return &GetDanmakusLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetDanmakusLogic) GetDanmakus(in *interact.GetDanmakusReq) (*interact.GetDanmakusResp, error) {
	return &interact.GetDanmakusResp{}, status.Error(codes.Unimplemented, "需要数据库 danmakus 表")
}
