// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"context"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"

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

func (l *LikeVideoLogic) LikeVideo() (resp *types.LikeResp, err error) {
	// todo: add your logic here and delete this line

	return
}
