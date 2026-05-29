// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"context"
	"net/http"
	"strconv"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/interact/interactclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCommentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCommentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCommentsLogic {
	return &ListCommentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListCommentsLogic) ListComments(r *http.Request) (resp *types.ListCommentsResp, err error) {
	videoId, _ := strconv.ParseInt(r.URL.Query().Get("video_id"), 10, 64)
	cursor, _ := strconv.ParseInt(r.URL.Query().Get("cursor"), 10, 64)
	sort := r.URL.Query().Get("sort")

	rpcResp, rpcErr := l.svcCtx.InteractClient.ListComments(l.ctx, &interactclient.ListCommentsReq{
		VideoId: videoId,
		Cursor:  cursor,
		Limit:   20,
		Sort:    sort,
	})
	if rpcErr != nil {
		return nil, rpcErr
	}

	comments := make([]types.CommentInfo, 0, len(rpcResp.Comments))
	for _, c := range rpcResp.Comments {
		comments = append(comments, types.CommentInfo{
			Id:         c.Id,
			UserId:     c.UserId,
			Username:   c.Username,
			Avatar:     c.Avatar,
			Content:    c.Content,
			LikeCount:  c.LikeCount,
			ReplyCount: int(c.ReplyCount),
			ParentId:   c.ParentId,
			CreatedAt:  c.CreatedAt,
		})
	}
	return &types.ListCommentsResp{
		Comments:   comments,
		NextCursor: rpcResp.NextCursor,
		HasMore:    rpcResp.HasMore,
	}, nil
}
