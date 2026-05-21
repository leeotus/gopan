// PostCommentLogic / ListCommentsLogic / DeleteCommentLogic — 评论相关逻辑。
package logic

import (
	"context"
	"database/sql"

	"gopan/rpc/interact/internal/svc"
	"gopan/rpc/interact/interact"
	"gopan/rpc/interact/store"

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
	commentId, err := l.svcCtx.InteractStore.InsertComment(l.ctx, in.UserId, in.VideoId, in.ParentId, in.Content)
	if err != nil {
		l.Logger.Errorf("insert comment error: %v", err)
		return nil, status.Error(codes.Internal, "评论失败")
	}
	return &interact.PostCommentResp{CommentId: commentId}, nil
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
	rows, err := l.svcCtx.InteractStore.ListComments(l.ctx, in.VideoId, in.Cursor, in.Limit, in.Sort)
	if err != nil {
		return nil, status.Error(codes.Internal, "查询失败")
	}

	hasMore := len(rows) > int(in.Limit)
	if hasMore {
		rows = rows[:in.Limit]
	}

	resp := &interact.ListCommentsResp{HasMore: hasMore}
	for _, r := range rows {
		resp.Comments = append(resp.Comments, &interact.CommentInfo{
			Id:         r.Id,
			UserId:     r.UserId,
			Content:    r.Content,
			LikeCount:  r.LikeCount,
			ReplyCount: int32(r.ReplyCount),
			ParentId:   r.ParentId,
			CreatedAt:  r.CreatedAt,
		})
		if !hasMore {
			resp.NextCursor = r.Id
		}
	}
	return resp, nil
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
	if err := l.svcCtx.InteractStore.DeleteComment(l.ctx, in.CommentId, in.UserId); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "评论不存在或无权删除")
		}
		l.Logger.Errorf("delete comment error: %v", err)
		return nil, status.Error(codes.Internal, "删除失败")
	}
	return &interact.DeleteCommentResp{}, nil
}

// 避免 import 未使用
var _ store.InteractStore
