package admin

import (
	"context"
	"strconv"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/admin/adminclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type RejectVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRejectVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RejectVideoLogic {
	return &RejectVideoLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *RejectVideoLogic) RejectVideo(videoId, adminId int64) (resp *types.BaseResp, err error) {
	_, rpcErr := l.svcCtx.AdminClient.RejectVideo(l.ctx, &adminclient.RejectVideoReq{
		VideoId: videoId,
		AdminId: adminId,
	})
	if rpcErr != nil {
		return &types.BaseResp{Message: rpcErr.Error()}, nil
	}
	return &types.BaseResp{Message: "ok"}, nil
}

var _ = strconv.Itoa
