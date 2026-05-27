package video

import (
	"context"
	"strconv"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/video/videoclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SavePlayProgressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSavePlayProgressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SavePlayProgressLogic {
	return &SavePlayProgressLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *SavePlayProgressLogic) SavePlayProgress(videoId, userId int64, position float64) (resp *types.BaseResp, err error) {
	_, rpcErr := l.svcCtx.VideoClient.SavePlayProgress(l.ctx, &videoclient.SavePlayProgressReq{
		VideoId:  videoId,
		UserId:   userId,
		Position: position,
	})
	if rpcErr != nil {
		return &types.BaseResp{Message: rpcErr.Error()}, nil
	}
	return &types.BaseResp{Message: "ok"}, nil
}

type GetPlayProgressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPlayProgressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPlayProgressLogic {
	return &GetPlayProgressLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *GetPlayProgressLogic) GetPlayProgress(videoIdStr, userIdStr string) (resp *types.BaseResp, err error) {
	videoId, _ := strconv.ParseInt(videoIdStr, 10, 64)
	userId, _ := strconv.ParseInt(userIdStr, 10, 64)

	r, rpcErr := l.svcCtx.VideoClient.GetPlayProgress(l.ctx, &videoclient.GetPlayProgressReq{
		VideoId: videoId,
		UserId:  userId,
	})
	if rpcErr != nil || r == nil {
		return &types.BaseResp{Message: "0.0"}, nil
	}
	return &types.BaseResp{Message: strconv.FormatFloat(r.Position, 'f', 1, 64)}, nil
}
