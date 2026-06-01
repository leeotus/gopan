package admin

import (
	"context"
	"strconv"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/admin/adminclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListVideosLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListVideosLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListVideosLogic {
	return &AdminListVideosLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AdminListVideosLogic) AdminListVideos(cursor int64, status int32) (resp *types.AdminListVideosResp, err error) {
	r, rpcErr := l.svcCtx.AdminClient.ListVideos(l.ctx, &adminclient.AdminListVideosReq{
		Cursor: cursor,
		Limit:  20,
		Status: status,
	})
	if rpcErr != nil {
		return nil, rpcErr
	}

	videos := make([]types.AdminVideoInfo, 0, len(r.Videos))
	for _, v := range r.Videos {
		videos = append(videos, types.AdminVideoInfo{
			Id:        v.Id,
			Title:     v.Title,
			CoverUrl:  v.CoverUrl,
			UserId:    v.UserId,
			Username:  v.Username,
			Status:    int(v.Status),
			PlayCount: v.PlayCount,
			CreatedAt: v.CreatedAt,
		})
	}

	return &types.AdminListVideosResp{
		Videos:     videos,
		NextCursor: r.NextCursor,
		HasMore:    r.HasMore,
	}, nil
}

// --- 类型兼容辅助 ---
var _ = strconv.Itoa
