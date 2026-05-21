// FavoriteLogic / UnfavoriteLogic / GetFavoriteStatusLogic — 收藏相关逻辑。
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
	if err := l.svcCtx.InteractStore.InsertFavorite(l.ctx, in.UserId, in.VideoId); err != nil {
		l.Logger.Errorf("insert favorite error: %v", err)
		return nil, status.Error(codes.Internal, "收藏失败")
	}
	return &interact.FavoriteResp{}, nil
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
	if err := l.svcCtx.InteractStore.DeleteFavorite(l.ctx, in.UserId, in.VideoId); err != nil {
		l.Logger.Errorf("delete favorite error: %v", err)
		return nil, status.Error(codes.Internal, "取消收藏失败")
	}
	return &interact.UnfavoriteResp{}, nil
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
	favorited, err := l.svcCtx.InteractStore.IsFavorited(l.ctx, in.UserId, in.VideoId)
	if err != nil {
		return nil, status.Error(codes.Internal, "查询失败")
	}
	return &interact.GetFavoriteStatusResp{Favorited: favorited}, nil
}
