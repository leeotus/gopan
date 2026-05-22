package video

import (
	"context"
	"net/http"
	"strings"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/video/videoclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewUploadStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadStatusLogic {
	return &UploadStatusLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UploadStatusLogic) SetRequest(r *http.Request) { l.r = r }

func (l *UploadStatusLogic) UploadStatus() (resp *types.UploadStatusResp, err error) {
	uploadId := ""
	if l.r != nil {
		uploadId = l.r.URL.Query().Get("upload_id")
	}
	if uploadId == "" {
		return &types.UploadStatusResp{}, nil
	}

	r, err := l.svcCtx.VideoClient.UploadStatus(l.ctx, &videoclient.UploadStatusReq{UploadId: uploadId})
	if err != nil {
		// gRPC 错误不 panic，返回空结构体
		return &types.UploadStatusResp{}, nil
	}

	return &types.UploadStatusResp{
		VideoId:        r.VideoId,
		TotalChunks:    r.TotalChunks,
		ReceivedChunks: r.ReceivedChunks,
	}, nil
}

// 编译检查
var _ = strings.Trim
