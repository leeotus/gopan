// GetVideoLogic 获取视频详情（含多码率转码信息）。
// 返回单个 video 的完整元数据 + 所有分辨率 HLS 地址。
package logic

import (
	"context"
	"database/sql"

	"gopan/rpc/video/internal/svc"
	"gopan/rpc/video/model"
	"gopan/rpc/video/video"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVideoLogic {
	return &GetVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetVideoLogic) GetVideo(in *video.GetVideoReq) (*video.GetVideoResp, error) {
	v, err := l.svcCtx.VideoStore.FindById(l.ctx, in.VideoId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "视频不存在")
		}
		return nil, status.Error(codes.Internal, "查询失败")
	}

	info := toVideoInfo(v)
	// 附加转码信息（多分辨率 HLS 地址）
	transcodes, _ := l.svcCtx.VideoStore.GetTranscodes(l.ctx, v.Id)
	for _, t := range transcodes {
		info.Transcodes = append(info.Transcodes, &video.TranscodeInfo{
			Resolution: t.Resolution,
			M3U8Url:    t.M3U8Url,
			Bitrate:    t.Bitrate,
		})
	}

	return &video.GetVideoResp{Video: info}, nil
}

// toVideoInfo 将 model.Video 转换为 proto 定义 of VideoInfo。
func toVideoInfo(m *model.Video) *video.VideoInfo {
	return &video.VideoInfo{
		Id:          m.Id,
		Title:       m.Title,
		UserId:      m.UserId,
		CoverUrl:    m.CoverUrl,
		PlayCount:   m.PlayCount,
		LikeCount:   m.LikeCount,
		Duration:    m.Duration,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt.Unix(),
		Description: m.Description,
		Category:    m.Category,
	}
}
