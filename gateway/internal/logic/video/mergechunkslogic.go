package video

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/video/videoclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type MergeChunksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewMergeChunksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MergeChunksLogic {
	return &MergeChunksLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *MergeChunksLogic) SetRequest(r *http.Request) { l.r = r }

func (l *MergeChunksLogic) MergeChunks() (resp *types.BaseResp, err error) {
	if l.r == nil {
		return &types.BaseResp{Message: "no request"}, nil
	}

	body, _ := io.ReadAll(l.r.Body)
	var req struct {
		VideoId  int64  `json:"video_id"`
		UploadId string `json:"upload_id"`
	}
	json.Unmarshal(body, &req)

	if req.VideoId == 0 || req.UploadId == "" {
		return &types.BaseResp{Message: "缺少 video_id 或 upload_id"}, nil
	}

	_, rpcErr := l.svcCtx.VideoClient.MergeChunks(l.ctx, &videoclient.MergeChunksReq{
		VideoId:  req.VideoId,
		UploadId: req.UploadId,
	})
	if rpcErr != nil {
		return &types.BaseResp{Message: rpcErr.Error()}, nil
	}

	return &types.BaseResp{Message: "merge completed"}, nil
}
