// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"context"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type InitUploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInitUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitUploadLogic {
	return &InitUploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InitUploadLogic) InitUpload(req *types.InitUploadReq) (resp *types.InitUploadResp, err error) {
	// todo: add your logic here and delete this line

	return
}
