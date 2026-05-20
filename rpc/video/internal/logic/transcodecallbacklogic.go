// TranscodeCallbackLogic 接收 transcode-svc 的转码完成回调。
// 更新视频状态、封面、时长，并写入各分辨率 HLS 地址。
package logic

import (
	"context"

	"gopan/rpc/video/internal/svc"
	"gopan/rpc/video/model"
	"gopan/rpc/video/video"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TranscodeCallbackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTranscodeCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TranscodeCallbackLogic {
	return &TranscodeCallbackLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TranscodeCallbackLogic) TranscodeCallback(in *video.TranscodeCallbackReq) (*video.TranscodeCallbackResp, error) {
	// status=2 表示转码成功，其他为失败
	vidStatus := int32(3)
	if in.Status == 2 {
		vidStatus = 2
	}

	if err := l.svcCtx.VideoStore.UpdateTranscode(l.ctx, in.VideoId, vidStatus, in.CoverUrl, in.Duration); err != nil {
		return nil, status.Error(codes.Internal, "更新转码状态失败")
	}

	// 写入各分辨率转码结果（360p / 480p / 720p / 1080p）
	for _, t := range in.Transcodes {
		l.svcCtx.VideoStore.InsertTranscode(l.ctx, &model.Transcode{
			VideoId:    in.VideoId,
			Resolution: t.Resolution,
			M3U8Url:    t.M3U8Url,
			Bitrate:    t.Bitrate,
		})
	}

	return &video.TranscodeCallbackResp{}, nil
}
