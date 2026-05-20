// PostCommentLogic / ListCommentsLogic / DeleteCommentLogic — 评论相关逻辑。
// TODO: 集成 comments 表。
package logic

import (
	"context"

	"gopan/rpc/interact/internal/svc"
	"gopan/rpc/interact/interact"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PostCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPostCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PostCommentLogic {
	return &PostCommentLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *PostCommentLogic) PostComment(in *interact.PostCommentReq) (*interact.PostCommentResp, error) {
	l.Logger.Infof("post comment: userId=%d, videoId=%d", in.UserId, in.VideoId)
	return &interact.PostCommentResp{}, status.Error(codes.Unimplemented, "需要数据库 comments 表")
}

type ListCommentsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCommentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCommentsLogic {
	return &ListCommentsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListCommentsLogic) ListComments(in *interact.ListCommentsReq) (*interact.ListCommentsResp, error) {
	return &interact.ListCommentsResp{}, status.Error(codes.Unimplemented, "需要数据库 comments 表")
}

type DeleteCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *DeleteCommentLogic) DeleteComment(in *interact.DeleteCommentReq) (*interact.DeleteCommentResp, error) {
	l.Logger.Infof("delete comment: commentId=%d, userId=%d", in.CommentId, in.UserId)
	return &interact.DeleteCommentResp{}, status.Error(codes.Unimplemented, "需要数据库 comments 表")
}
