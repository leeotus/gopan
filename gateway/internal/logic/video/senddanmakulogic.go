// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"context"
	"net/http"
	"strconv"

	"gopan/gateway/internal/middleware"
	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/interact/interactclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendDanmakuLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendDanmakuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendDanmakuLogic {
	return &SendDanmakuLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendDanmakuLogic) SendDanmaku(r *http.Request, req *types.DanmakuMsg) (resp *types.BaseResp, err error) {
	videoId, _ := strconv.ParseInt(r.URL.Query().Get("video_id"), 10, 64)
	userId := middleware.GetUserIdFromContext(l.ctx)

	_, rpcErr := l.svcCtx.InteractClient.SendDanmaku(l.ctx, &interactclient.SendDanmakuReq{
		UserId:  userId,
		VideoId: videoId,
		Content: req.Content,
		Time:    req.Time,
		Color:   req.Color,
		Mode:    int32(req.Mode),
	})
	if rpcErr != nil {
		return &types.BaseResp{Message: rpcErr.Error()}, nil
	}
	return &types.BaseResp{Message: "ok"}, nil
}
