package logic

import (
	"context"

	"gopan/rpc/admin/admin"
	"gopan/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteVideoLogic {
	return &DeleteVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteVideoLogic) DeleteVideo(in *admin.AdminDeleteVideoReq) (*admin.AdminDeleteVideoResp, error) {
	// todo: add your logic here and delete this line

	return &admin.AdminDeleteVideoResp{}, nil
}
