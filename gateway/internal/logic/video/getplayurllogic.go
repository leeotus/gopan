// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"context"
	"net/http"
	"strconv"

	"gopan/gateway/internal/middleware"
	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/stream/streamclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPlayUrlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPlayUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPlayUrlLogic {
	return &GetPlayUrlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPlayUrlLogic) GetPlayUrl(r *http.Request, req *types.PlayUrlReq) (resp *types.PlayUrlResp, err error) {
	videoId, _ := strconv.ParseInt(r.URL.Query().Get("video_id"), 10, 64)
	userId := middleware.GetUserIdFromContext(l.ctx)

	rpcResp, rpcErr := l.svcCtx.StreamClient.GetPlayUrl(l.ctx, &streamclient.GetPlayUrlReq{
		VideoId:    videoId,
		UserId:     userId,
		Resolution: req.Resolution,
	})
	if rpcErr != nil {
		return nil, rpcErr
	}

	streams := make([]types.PlayStream, 0, len(rpcResp.Streams))
	for _, s := range rpcResp.Streams {
		streams = append(streams, types.PlayStream{
			Resolution: s.Resolution,
			Url:        s.Url,
			Bitrate:    int(s.Bitrate),
		})
	}
	return &types.PlayUrlResp{
		M3u8Url: rpcResp.M3U8Url,
		Streams: streams,
	}, nil
}
