package video

import (
	"context"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/interact/interactclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDanmakusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDanmakusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDanmakusLogic {
	return &GetDanmakusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDanmakusLogic) GetDanmakus(videoId int64, timeVal float64) (resp *types.GetDanmakusResp, err error) {
	rpcResp, rpcErr := l.svcCtx.InteractClient.GetDanmakus(l.ctx, &interactclient.GetDanmakusReq{
		VideoId: videoId,
		Time:    timeVal,
	})
	if rpcErr != nil {
		return nil, rpcErr
	}

	danmakus := make([]types.DanmakuMsg, 0, len(rpcResp.Danmakus))
	for _, d := range rpcResp.Danmakus {
		danmakus = append(danmakus, types.DanmakuMsg{
			Content: d.Content,
			Time:    d.Time,
			Color:   d.Color,
			Mode:    int(d.Mode),
		})
	}
	return &types.GetDanmakusResp{Danmakus: danmakus}, nil
}
