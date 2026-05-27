// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package admin

import (
	"context"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListVideosLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListVideosLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListVideosLogic {
	return &AdminListVideosLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListVideosLogic) AdminListVideos() (resp *types.BaseResp, err error) {
	// todo: add your logic here and delete this line

	return
}
