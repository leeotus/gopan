// MergeChunksLogic 检查完整性 → 发 Kafka 异步合并 → 立即返回。
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

	count, err := l.svcCtx.UploadProgress.CountReceived(l.ctx, in.UploadId)
	if err != nil {
		l.Logger.Errorf("CountReceived error: %v", err)
		count = 0
	}

	if int32(count) < v.TotalChunks {
		return &video.MergeChunksResp{Status: "incomplete"}, nil
	}

	// 构造 chunk key 列表
	prefix := fmt.Sprintf("parts/%d", in.VideoId)
	chunkKeys := make([]string, v.TotalChunks)
	for i := int32(0); i < v.TotalChunks; i++ {
		chunkKeys[i] = fmt.Sprintf(prefix+"/chunk_%d", i)
	}

	// 发 Kafka 异步合并任务
	task := commonkafka.MergeTask{
		VideoId:   in.VideoId,
		UploadId:  in.UploadId,
		ChunkKeys: chunkKeys,
		TotalChunks: v.TotalChunks,
	}
	body, _ := json.Marshal(task)
	err = l.svcCtx.KafkaMergeWriter.WriteMessages(l.ctx, kafkago.Message{
		Key:   []byte(fmt.Sprintf("merge-video-%d", in.VideoId)),
		Value: body,
	})
	if err != nil {
		l.Logger.Errorf("kafka write merge task error: %v", err)
		return nil, status.Error(codes.Internal, "提交合并任务失败")
	}

	l.Logger.Infof("kafka merge task sent: video_id=%d chunks=%d", in.VideoId, v.TotalChunks)

	return &video.MergeChunksResp{Status: "complete", ObjectKey: fmt.Sprintf("videos/%d/source.mp4", in.VideoId)}, nil
}
