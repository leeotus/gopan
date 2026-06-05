// SendDanmakuLogic / GetDanmakusLogic — 弹幕相关逻辑。
package logic

import (
	"context"
	"fmt"

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
	danmakuId, err := l.svcCtx.InteractStore.InsertDanmaku(
		l.ctx, in.UserId, in.VideoId, in.Content, in.Time, in.Color, in.Mode,
	)
	if err != nil {
		l.Logger.Errorf("insert danmaku error: %v", err)
		return nil, status.Error(codes.Internal, "发送弹幕失败")
	}

	// Redis Pub/Sub 实时推送
	message := fmt.Sprintf(`{"id":%d,"user_id":%d,"content":"%s","time":%.1f,"color":"%s","mode":%d}`,
		danmakuId, in.UserId, in.Content, in.Time, in.Color, in.Mode)
	if err := l.svcCtx.Redis.Publish(l.ctx, fmt.Sprintf("danmaku:%d", in.VideoId), message).Err(); err != nil {
		l.Logger.Errorf("redis publish danmaku error: %v", err)
	}

	return &interact.SendDanmakuResp{DanmakuId: danmakuId}, nil
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
	rows, err := l.svcCtx.InteractStore.GetDanmakus(l.ctx, in.VideoId, in.Time)
	if err != nil {
		return nil, status.Error(codes.Internal, "查询弹幕失败")
	}

	resp := &interact.GetDanmakusResp{}
	for _, r := range rows {
		resp.Danmakus = append(resp.Danmakus, &interact.DanmakuInfo{
			Id:      r.Id,
			UserId:  r.UserId,
			Content: r.Content,
			Time:    r.Time,
			Color:   r.Color,
			Mode:    r.Mode,
		})
	}
	return resp, nil
}
