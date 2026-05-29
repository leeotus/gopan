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

type UnlikeVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUnlikeVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlikeVideoLogic {
	return &UnlikeVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnlikeVideoLogic) UnlikeVideo(r *http.Request) (resp *types.LikeResp, err error) {
	videoId, _ := strconv.ParseInt(r.URL.Query().Get("video_id"), 10, 64)
	userId := middleware.GetUserIdFromContext(l.ctx)

	rpcResp, rpcErr := l.svcCtx.InteractClient.Unlike(l.ctx, &interactclient.UnlikeReq{
		UserId:  userId,
		VideoId: videoId,
	})
	if rpcErr != nil {
		return nil, rpcErr
	}
	return &types.LikeResp{
		LikeCount: rpcResp.LikeCount,
		Liked:     false,
	}, nil
}
