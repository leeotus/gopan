// FavoriteLogic / UnfavoriteLogic / GetFavoriteStatusLogic — 收藏相关逻辑。
// TODO: 集成 favorites 表。
package logic

import (
	"context"

	"gopan/rpc/interact/internal/svc"
	"gopan/rpc/interact/interact"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FavoriteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFavoriteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FavoriteLogic {
	return &FavoriteLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *FavoriteLogic) Favorite(in *interact.FavoriteReq) (*interact.FavoriteResp, error) {
	l.Logger.Infof("favorite: userId=%d, videoId=%d", in.UserId, in.VideoId)
	return &interact.FavoriteResp{}, status.Error(codes.Unimplemented, "需要数据库 favorites 表")
}

type UnfavoriteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnfavoriteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfavoriteLogic {
	return &UnfavoriteLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *UnfavoriteLogic) Unfavorite(in *interact.UnfavoriteReq) (*interact.UnfavoriteResp, error) {
	l.Logger.Infof("unfavorite: userId=%d, videoId=%d", in.UserId, in.VideoId)
	return &interact.UnfavoriteResp{}, status.Error(codes.Unimplemented, "需要数据库 favorites 表")
}

type GetFavoriteStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFavoriteStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFavoriteStatusLogic {
	return &GetFavoriteStatusLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetFavoriteStatusLogic) GetFavoriteStatus(in *interact.GetFavoriteStatusReq) (*interact.GetFavoriteStatusResp, error) {
	return &interact.GetFavoriteStatusResp{}, nil
}
