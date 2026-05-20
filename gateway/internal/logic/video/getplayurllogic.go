// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"context"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPlayUrlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPlayUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPlayUrlLogic {
	return &GetPlayUrlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPlayUrlLogic) GetPlayUrl(req *types.PlayUrlReq) (resp *types.PlayUrlResp, err error) {
	// todo: add your logic here and delete this line

	return
}
