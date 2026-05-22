package video

import (
	"context"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/video/videoclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadStatusLogic {
	return &UploadStatusLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UploadStatusLogic) UploadStatus() (resp *types.BaseResp, err error) {
	// TODO: 从 query param 取 upload_id
	// r, _ := l.svcCtx.VideoClient.UploadStatus(l.ctx, &videoclient.UploadStatusReq{UploadId: uploadId})
	return &types.BaseResp{Message: "upload-status OK"}, nil
}

var _ videoclient.Video = nil
