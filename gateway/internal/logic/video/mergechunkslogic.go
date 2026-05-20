// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"context"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MergeChunksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMergeChunksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MergeChunksLogic {
	return &MergeChunksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MergeChunksLogic) MergeChunks() (resp *types.BaseResp, err error) {
	// todo: add your logic here and delete this line

	return
}
