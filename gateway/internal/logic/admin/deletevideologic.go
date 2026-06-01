package admin

import (
	"context"
	"strconv"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/admin/adminclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteVideoLogic {
	return &DeleteVideoLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DeleteVideoLogic) DeleteVideo(videoId, adminId int64) (resp *types.BaseResp, err error) {
	_, rpcErr := l.svcCtx.AdminClient.DeleteVideo(l.ctx, &adminclient.AdminDeleteVideoReq{
		VideoId: videoId,
		AdminId: adminId,
	})
	if rpcErr != nil {
		return &types.BaseResp{Message: rpcErr.Error()}, nil
	}
	return &types.BaseResp{Message: "ok"}, nil
}

var _ = strconv.Itoa
