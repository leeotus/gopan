// MergeChunksLogic 检查完整性 → MinIO ComposeObject → 发送 Kafka → 清理。
package logic

import (
	"context"
	"encoding/json"
	"fmt"

	commonkafka "gopan/common/kafka"
	"gopan/rpc/video/internal/svc"
	"gopan/rpc/video/video"

	kafkago "github.com/segmentio/kafka-go"
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
	return &MergeChunksLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *MergeChunksLogic) MergeChunks(in *video.MergeChunksReq) (*video.MergeChunksResp, error) {
	v, err := l.svcCtx.VideoStore.FindById(l.ctx, in.VideoId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "视频不存在")
	}

	// 1. 校验完整性：已收到的 chunk 数量 = total_chunks
	count, err := l.svcCtx.UploadProgress.CountReceived(l.ctx, in.UploadId)
	if err != nil {
		count = 0
	}

	if int32(count) < v.TotalChunks {
		// 缺失分片：返回缺失列表
		received, _ := l.svcCtx.UploadProgress.GetReceived(l.ctx, in.UploadId)
		receivedSet := make(map[int32]bool)
		for _, r := range received {
			receivedSet[r] = true
		}
		var missing []int32
		for i := int32(0); i < v.TotalChunks; i++ {
			if !receivedSet[i] {
				missing = append(missing, i)
			}
		}
		return &video.MergeChunksResp{Status: "incomplete", MissingChunks: missing}, nil
	}

	// 2. 合并：MinIO ComposeObject
	prefix := fmt.Sprintf("parts/%d", in.VideoId)
	var sourceKeys []string
	for i := int32(0); i < v.TotalChunks; i++ {
		sourceKeys = append(sourceKeys, fmt.Sprintf(prefix+"/chunk_%d", i))
	}
	destKey := fmt.Sprintf("videos/%d/source.mp4", in.VideoId)
	if err := l.svcCtx.MinioClient.ComposeObject(l.ctx, destKey, sourceKeys); err != nil {
		l.Logger.Errorf("minio compose error: %v", err)
		return nil, status.Error(codes.Internal, "分片合并失败")
	}

	// 3. 更新视频状态为"转码中"
	_ = l.svcCtx.VideoStore.UpdateStatus(l.ctx, in.VideoId, 1)
	_ = l.svcCtx.UploadProgress.Clear(l.ctx, in.UploadId)

	// 4. 发送 Kafka 转码任务
	task := commonkafka.TranscodeTask{VideoId: in.VideoId, ObjectKey: destKey}
	body, _ := json.Marshal(task)
	key := []byte(fmt.Sprintf("video-%d", in.VideoId))
	err = l.svcCtx.KafkaWriter.WriteMessages(l.ctx, kafkago.Message{Key: key, Value: body})
	if err != nil {
		l.Logger.Errorf("kafka write transcode task error: %v", err)
	} else {
		l.Logger.Infof("kafka transcode task sent: video_id=%d", in.VideoId)
	}

	return &video.MergeChunksResp{Status: "complete", ObjectKey: destKey}, nil
}
