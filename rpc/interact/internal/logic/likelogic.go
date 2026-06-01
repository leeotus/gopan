// LikeLogic / UnlikeLogic / GetLikeStatusLogic — 点赞相关逻辑。
package logic

import (
	"context"
	"database/sql"

	"gopan/rpc/interact/internal/svc"
	"gopan/rpc/interact/interact"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LikeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeLogic {
	return &LikeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *LikeLogic) Like(in *interact.LikeReq) (*interact.LikeResp, error) {
	if err := l.svcCtx.InteractStore.InsertLike(l.ctx, in.UserId, in.VideoId); err != nil {
		l.Logger.Errorf("insert like error: %v", err)
		return nil, status.Error(codes.Internal, "点赞失败")
	}
	count := l.svcCtx.InteractStore.CountLikes(l.ctx, in.VideoId)
	return &interact.LikeResp{LikeCount: count}, nil
}

type UnlikeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnlikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlikeLogic {
	return &UnlikeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *UnlikeLogic) Unlike(in *interact.UnlikeReq) (*interact.UnlikeResp, error) {
	if err := l.svcCtx.InteractStore.DeleteLike(l.ctx, in.UserId, in.VideoId); err != nil {
		if err == sql.ErrNoRows {
			return &interact.UnlikeResp{}, nil
		}
		l.Logger.Errorf("delete like error: %v", err)
		return nil, status.Error(codes.Internal, "取消点赞失败")
	}
	return &interact.UnlikeResp{}, nil
}

type GetLikeStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLikeStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLikeStatusLogic {
	return &GetLikeStatusLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetLikeStatusLogic) GetLikeStatus(in *interact.GetLikeStatusReq) (*interact.GetLikeStatusResp, error) {
	liked, err := l.svcCtx.InteractStore.IsLiked(l.ctx, in.UserId, in.VideoId)
	if err != nil {
		return nil, status.Error(codes.Internal, "查询失败")
	}
	return &interact.GetLikeStatusResp{Liked: liked}, nil
}
