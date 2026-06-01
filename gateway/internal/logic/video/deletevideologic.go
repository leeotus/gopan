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
	"gopan/rpc/video/videoclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteVideoLogic {
	return &DeleteVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteVideoLogic) DeleteVideo(r *http.Request) (resp *types.BaseResp, err error) {
	videoId, _ := strconv.ParseInt(r.URL.Query().Get("video_id"), 10, 64)
	userId := middleware.GetUserIdFromContext(l.ctx)

	_, rpcErr := l.svcCtx.VideoClient.DeleteVideo(l.ctx, &videoclient.DeleteVideoReq{
		VideoId: videoId,
		UserId:  userId,
	})
	if rpcErr != nil {
		return &types.BaseResp{Message: rpcErr.Error()}, nil
	}
	return &types.BaseResp{Message: "ok"}, nil
}
