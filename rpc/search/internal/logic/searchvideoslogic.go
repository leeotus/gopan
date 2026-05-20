// SearchVideosLogic / IndexVideoLogic / RemoveVideoLogic — 搜索相关逻辑。
// TODO: 集成 Elasticsearch 客户端，实现全文搜索和索引管理。
package logic

import (
	"context"

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

// SearchVideos 全文搜索视频。
// TODO: 调用 ES multi_match 查询 title + description 字段。
func (l *SearchVideosLogic) SearchVideos(in *search.SearchVideosReq) (*search.SearchVideosResp, error) {
	l.Logger.Infof("search videos: keyword=%s, page=%d", in.Keyword, in.Page)
	return &search.SearchVideosResp{}, status.Error(codes.Unimplemented, "需要配置 Elasticsearch")
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
	l.Logger.Infof("index video: videoId=%d, title=%s", in.VideoId, in.Title)
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
	l.Logger.Infof("remove video from index: videoId=%d", in.VideoId)
	return &search.RemoveVideoResp{}, nil
}
