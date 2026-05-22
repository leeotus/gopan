package video

import (
	"context"
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

func (l *MergeChunksLogic) MergeChunks() (resp *types.BaseResp, err error) {
	if l.r == nil {
		return &types.BaseResp{Message: "no request"}, nil
	}

	body, _ := io.ReadAll(l.r.Body)
	// 简单解析 JSON body: {"video_id": xxx, "upload_id": "xxx"}
	// 当前直接返回 ok，后续完善

	_ = body
	_ = videoclient.NewVideo

	return &types.BaseResp{Message: "merge OK"}, nil
}
