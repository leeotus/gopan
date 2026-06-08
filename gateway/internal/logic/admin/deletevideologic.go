package admin

import (
	"context"
	"errors"

	"gopan/gateway/internal/middleware"
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
	if middleware.GetRoleFromContext(l.ctx) != 1 {
		return &types.BaseResp{Message: "权限不足：仅管理员可操作"}, errors.New("permission denied")
	}

	_, rpcErr := l.svcCtx.AdminClient.DeleteVideo(l.ctx, &adminclient.AdminDeleteVideoReq{
		VideoId: videoId,
		AdminId: adminId,
	})
	if rpcErr != nil {
		return &types.BaseResp{Message: rpcErr.Error()}, nil
	}
	return &types.BaseResp{Message: "ok"}, nil
}
