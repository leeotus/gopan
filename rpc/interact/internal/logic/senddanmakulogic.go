// SendDanmakuLogic / GetDanmakusLogic — 弹幕相关逻辑。
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
	danmakuId, err := l.svcCtx.InteractStore.InsertDanmaku(
		l.ctx, in.UserId, in.VideoId, in.Content, in.Time, in.Color, in.Mode,
	)
	if err != nil {
		l.Logger.Errorf("insert danmaku error: %v", err)
		return nil, status.Error(codes.Internal, "发送弹幕失败")
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
