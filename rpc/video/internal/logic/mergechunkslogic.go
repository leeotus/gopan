// MergeChunksLogic 检查完整性 → 流式合并分片 → 发送 Kafka → 清理。
package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

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

	count, err := l.svcCtx.UploadProgress.CountReceived(l.ctx, in.UploadId)
	if err != nil {
		l.Logger.Errorf("CountReceived error: %v", err)
		count = 0
	}

	if int32(count) < v.TotalChunks {
		received, _ := l.svcCtx.UploadProgress.GetReceived(l.ctx, in.UploadId)
		receivedSet := make(map[int32]bool)
		for _, r := range received { receivedSet[r] = true }
		var missing []int32
		for i := int32(0); i < v.TotalChunks; i++ {
			if !receivedSet[i] { missing = append(missing, i) }
		}
		return &video.MergeChunksResp{Status: "incomplete", MissingChunks: missing}, nil
	}

	// 流式合并：从 MinIO 逐个下载分片 → 拼接 → 上传完整文件
	prefix := fmt.Sprintf("parts/%d", in.VideoId)
	var buf bytes.Buffer
	for i := int32(0); i < v.TotalChunks; i++ {
		key := fmt.Sprintf(prefix+"/chunk_%d", i)
		reader, err := l.svcCtx.MinioClient.GetObject(l.ctx, key)
		if err != nil {
			l.Logger.Errorf("minio get chunk error: key=%s err=%v", key, err)
			return nil, status.Error(codes.Internal, "读取分片失败")
		}
		if _, err := io.Copy(&buf, reader); err != nil {
			reader.Close()
			return nil, status.Error(codes.Internal, "读取分片失败")
		}
		reader.Close()
	}

	destKey := fmt.Sprintf("videos/%d/source.mp4", in.VideoId)
	if err := l.svcCtx.MinioClient.PutObject(l.ctx, destKey, bytes.NewReader(buf.Bytes()), int64(buf.Len()), "video/mp4"); err != nil {
		l.Logger.Errorf("minio put merged video error: %v", err)
		return nil, status.Error(codes.Internal, "合并视频失败")
	}
	l.Logger.Infof("merge completed: video_id=%d dest=%s size=%d", in.VideoId, destKey, buf.Len())

	_ = l.svcCtx.VideoStore.UpdateStatus(l.ctx, in.VideoId, 1)
	_ = l.svcCtx.UploadProgress.Clear(l.ctx, in.UploadId)

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
