// MergeChunksLogic 合并上传分片（暂为桩实现，需集成 MinIO 后才完整）。
package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"gopan/rpc/video/internal/svc"
	"gopan/rpc/video/video"

	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MergeChunksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMergeChunksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MergeChunksLogic {
	return &MergeChunksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// transcodeTask 发送到 Kafka 的转码任务消息体。
// 简化版：不含 resolutions，默认输出 1080p HLS。
type transcodeTask struct {
	VideoId   int64  `json:"video_id"`
	ObjectKey string `json:"object_key"`
}

func (l *MergeChunksLogic) MergeChunks(in *video.MergeChunksReq) (*video.MergeChunksResp, error) {
	// 合并完成后更新视频状态为"转码中"
	if err := l.svcCtx.VideoStore.UpdateStatus(l.ctx, in.VideoId, 1); err != nil {
		return nil, status.Error(codes.Internal, "更新视频状态失败")
	}

	v, err := l.svcCtx.VideoStore.FindById(l.ctx, in.VideoId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "视频不存在")
	}

	// --- 发送 Kafka 转码任务 ---
	task := transcodeTask{
		VideoId:   in.VideoId,
		ObjectKey: v.ObjectKey,
	}
	body, _ := json.Marshal(task)
	key := []byte(fmt.Sprintf("video-%d", in.VideoId)) // 相同 video_id 进入同一个 partition，保证顺序

	err = l.svcCtx.KafkaWriter.WriteMessages(l.ctx, kafka.Message{
		Key:   key,
		Value: body,
	})
	if err != nil {
		l.Logger.Errorf("kafka write transcode task error: %v", err)
		// 不阻塞返回，转码失败可后续重试
	} else {
		l.Logger.Infof("kafka transcode task sent: video_id=%d", in.VideoId)
	}

	return &video.MergeChunksResp{
		ObjectKey: v.ObjectKey,
		FileHash:  v.FileHash,
	}, nil
}
