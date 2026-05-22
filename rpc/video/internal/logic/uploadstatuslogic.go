// UploadStatusLogic 查询上传进度（已收到的 chunk 列表）。
package logic

import (
	"context"
	"database/sql"

	"gopan/rpc/video/internal/svc"
	"gopan/rpc/video/video"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UploadStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUploadStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadStatusLogic {
	return &UploadStatusLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *UploadStatusLogic) UploadStatus(in *video.UploadStatusReq) (*video.UploadStatusResp, error) {
	// 1. 查 MySQL 获取 total_chunks + video_id
	v, err := l.svcCtx.VideoStore.FindByUploadId(l.ctx, in.UploadId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "上传会话不存在")
		}
		return nil, status.Error(codes.Internal, "查询失败")
	}

	// 2. 查 Redis 获取已收到的 chunk 列表
	received, err := l.svcCtx.UploadProgress.GetReceived(l.ctx, in.UploadId)
	if err != nil {
		l.Logger.Errorf("redis get received error: %v", err)
		received = nil
	}
	l.Logger.Infof("uploadstatus: upload_id=%s, total_chunks=%d, received=%v", in.UploadId, v.TotalChunks, received)

	return &video.UploadStatusResp{
		VideoId:        v.Id,
		TotalChunks:    v.TotalChunks,
		ReceivedChunks: received,
	}, nil
}
