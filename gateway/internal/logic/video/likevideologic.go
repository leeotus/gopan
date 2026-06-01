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

type LikeVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLikeVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeVideoLogic {
	return &LikeVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LikeVideoLogic) LikeVideo(r *http.Request) (resp *types.LikeResp, err error) {
	videoId, _ := strconv.ParseInt(r.URL.Query().Get("video_id"), 10, 64)
	userId := middleware.GetUserIdFromContext(l.ctx)

	rpcResp, rpcErr := l.svcCtx.InteractClient.Like(l.ctx, &interactclient.LikeReq{
		UserId:  userId,
		VideoId: videoId,
	})
	if rpcErr != nil {
		return nil, rpcErr
	}
	return &types.LikeResp{
		LikeCount: rpcResp.LikeCount,
		Liked:     true,
	}, nil
}
