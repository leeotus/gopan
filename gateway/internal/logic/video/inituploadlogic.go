package video

import (
	"context"

	"gopan/gateway/internal/middleware"
	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/video/videoclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type InitUploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInitUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitUploadLogic {
	return &InitUploadLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *InitUploadLogic) InitUpload(req *types.InitUploadReq) (resp *types.InitUploadResp, err error) {
	userId := middleware.GetUserIdFromContext(l.ctx)

	r, err := l.svcCtx.VideoClient.InitUpload(l.ctx, &videoclient.InitUploadReq{
		Filename:    req.Filename,
		Title:       req.Title,
		FileSize:    req.FileSize,
		TotalChunks: req.TotalChunks,
		UserId:      userId,
	})
	if err != nil {
		return nil, err
	}

	return &types.InitUploadResp{
		VideoId:  r.VideoId,
		UploadId: r.UploadId,
	}, nil
}

var _ videoclient.Video = nil
