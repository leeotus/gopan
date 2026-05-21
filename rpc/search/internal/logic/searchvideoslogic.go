// search-svc logic 层：SearchVideos / IndexVideo / RemoveVideo
package logic

import (
	"context"

	"gopan/common/es"
	"gopan/rpc/search/internal/svc"
	"gopan/rpc/search/search"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SearchVideosLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchVideosLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchVideosLogic {
	return &SearchVideosLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *SearchVideosLogic) SearchVideos(in *search.SearchVideosReq) (*search.SearchVideosResp, error) {
	if l.svcCtx.ESClient == nil {
		return nil, status.Error(codes.Internal, "ES 未配置")
	}

	result, err := l.svcCtx.ESClient.SearchVideos(l.ctx, in.Keyword, in.Category, int(in.Page), int(in.Size))
	if err != nil {
		l.Logger.Errorf("es search error: %v", err)
		return nil, status.Error(codes.Internal, "搜索失败")
	}

	videos := make([]*search.SearchVideoInfo, 0, len(result.Hits))
	for _, h := range result.Hits {
		videos = append(videos, &search.SearchVideoInfo{
			Id:          h.VideoId,
			Title:       h.Title,
			CoverUrl:    h.CoverUrl,
			UserId:      h.UserId,
			Username:    h.Username,
			PlayCount:   h.PlayCount,
			LikeCount:   h.LikeCount,
			Description: h.Description,
			Duration:    h.Duration,
			Category:    h.Category,
			CreatedAt:   h.CreatedAt,
		})
	}

	return &search.SearchVideosResp{
		Videos: videos,
		Total:  result.Total,
		Page:   in.Page,
		Size:   in.Size,
	}, nil
}

type IndexVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIndexVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IndexVideoLogic {
	return &IndexVideoLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *IndexVideoLogic) IndexVideo(in *search.IndexVideoReq) (*search.IndexVideoResp, error) {
	if l.svcCtx.ESClient == nil {
		return &search.IndexVideoResp{}, nil
	}

	doc := &es.VideoDoc{
		VideoId:     in.VideoId,
		Title:       in.Title,
		Description: in.Description,
		Category:    in.Category,
		UserId:      in.UserId,
		Username:    in.Username,
		CoverUrl:    in.CoverUrl,
		PlayCount:   in.PlayCount,
		LikeCount:   in.LikeCount,
		Duration:    in.Duration,
	}
	if err := l.svcCtx.ESClient.IndexVideo(l.ctx, doc); err != nil {
		l.Logger.Errorf("es index error: %v", err)
		return &search.IndexVideoResp{}, nil // 索引失败不影响主流程
	}
	return &search.IndexVideoResp{}, nil
}

type RemoveVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveVideoLogic {
	return &RemoveVideoLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *RemoveVideoLogic) RemoveVideo(in *search.RemoveVideoReq) (*search.RemoveVideoResp, error) {
	if l.svcCtx.ESClient == nil {
		return &search.RemoveVideoResp{}, nil
	}
	if err := l.svcCtx.ESClient.RemoveVideo(l.ctx, in.VideoId); err != nil {
		l.Logger.Errorf("es remove error: %v", err)
	}
	return &search.RemoveVideoResp{}, nil
}
