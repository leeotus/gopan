package admin

import (
	"context"
	"strconv"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/admin/adminclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApproveVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApproveVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveVideoLogic {
	return &ApproveVideoLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ApproveVideoLogic) ApproveVideo(videoId, adminId int64) (resp *types.BaseResp, err error) {
	_, rpcErr := l.svcCtx.AdminClient.ApproveVideo(l.ctx, &adminclient.ApproveVideoReq{
		VideoId: videoId,
		AdminId: adminId,
	})
	if rpcErr != nil {
		return &types.BaseResp{Message: rpcErr.Error()}, nil
	}
	return &types.BaseResp{Message: "ok"}, nil
}

var _ = strconv.Itoa
