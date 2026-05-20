// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"context"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PostCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPostCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PostCommentLogic {
	return &PostCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PostCommentLogic) PostComment(req *types.PostCommentReq) (resp *types.BaseResp, err error) {
	// todo: add your logic here and delete this line

	return
}
