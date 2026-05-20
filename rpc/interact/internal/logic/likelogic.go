// LikeLogic / UnlikeLogic / GetLikeStatusLogic — 点赞相关逻辑。
// TODO: 集成点赞表和 Redis 计数器。
package logic

import (
	"context"

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
	l.Logger.Infof("like: userId=%d, videoId=%d", in.UserId, in.VideoId)
	return &interact.LikeResp{}, status.Error(codes.Unimplemented, "需要数据库 likes 表")
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
	l.Logger.Infof("unlike: userId=%d, videoId=%d", in.UserId, in.VideoId)
	return &interact.UnlikeResp{}, status.Error(codes.Unimplemented, "需要数据库 likes 表")
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
	return &interact.GetLikeStatusResp{}, nil
}
