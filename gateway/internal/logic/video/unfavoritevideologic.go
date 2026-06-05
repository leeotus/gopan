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

type UnfavoriteVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUnfavoriteVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfavoriteVideoLogic {
	return &UnfavoriteVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnfavoriteVideoLogic) UnfavoriteVideo(r *http.Request) (resp *types.BaseResp, err error) {
	videoId, _ := strconv.ParseInt(r.URL.Query().Get("video_id"), 10, 64)
	userId := middleware.GetUserIdFromContext(l.ctx)

	_, rpcErr := l.svcCtx.InteractClient.Unfavorite(l.ctx, &interactclient.UnfavoriteReq{
		UserId:  userId,
		VideoId: videoId,
	})
	if rpcErr != nil {
		return &types.BaseResp{Message: rpcErr.Error()}, nil
	}
	return &types.BaseResp{Message: "ok"}, nil
}
