// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package search

import (
	"context"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/search/search"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchLogic {
	return &SearchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchLogic) Search(req *types.SearchReq) (resp *types.SearchResp, err error) {
	rpcResp, err := l.svcCtx.SearchClient.SearchVideos(l.ctx, &search.SearchVideosReq{
		Keyword:  req.Keyword,
		Category: req.Category,
		Page:     int32(req.Page),
		Size:     int32(req.Size),
	})
	if err != nil {
		return nil, err
	}

	videos := make([]types.SearchVideoInfo, 0, len(rpcResp.Videos))
	for _, v := range rpcResp.Videos {
		videos = append(videos, types.SearchVideoInfo{
			Id:          v.Id,
			Title:       v.Title,
			CoverUrl:    v.CoverUrl,
			UserId:      v.UserId,
			Username:    v.Username,
			PlayCount:   v.PlayCount,
			LikeCount:   v.LikeCount,
			Description: v.Description,
			Duration:    int(v.Duration),
			Category:    v.Category,
			CreatedAt:   v.CreatedAt,
			Score:       v.Score,
		})
	}

	return &types.SearchResp{
		Videos: videos,
		Total:  rpcResp.Total,
		Page:   int(rpcResp.Page),
		Size:   int(rpcResp.Size),
	}, nil
}
