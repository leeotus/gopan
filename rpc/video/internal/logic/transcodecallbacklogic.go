// TranscodeCallbackLogic 接收 transcode-svc 的转码完成回调。
// 更新视频状态、封面、时长，并写入各分辨率 HLS 地址。
// 转码成功后会做两件事：① 异步写 ES 索引（搜索可用）② 投递 Kafka SummaryTask 触发 AI 摘要离线生成。
package logic

import (
	"context"
	"encoding/json"
	"fmt"

	commonkafka "gopan/common/kafka"
	"gopan/rpc/search/searchclient"
	"gopan/rpc/video/internal/svc"
	"gopan/rpc/video/model"
	"gopan/rpc/video/video"

	kafkago "github.com/segmentio/kafka-go"
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

	// 写入各分辨率转码结果
	for _, t := range in.Transcodes {
		l.svcCtx.VideoStore.InsertTranscode(l.ctx, &model.Transcode{
			VideoId:    in.VideoId,
			Resolution: t.Resolution,
			M3U8Url:    t.M3U8Url,
			Bitrate:    t.Bitrate,
		})
	}

	// 转码成功后续工作
	if vidStatus == 2 {
		v, _ := l.svcCtx.VideoStore.FindById(l.ctx, in.VideoId)
		if v != nil {
			// 1. 写入 ES 索引（异步，失败不影响主流程）
			go func() {
				_, _ = l.svcCtx.SearchClient.IndexVideo(context.Background(), &searchclient.IndexVideoReq{
					VideoId:     v.Id,
					Title:       v.Title,
					Description: v.Description,
					Category:    v.Category,
					UserId:      v.UserId,
					Username:    "", // TODO: 从 user-svc 获取
					CoverUrl:    v.CoverUrl,
					PlayCount:   v.PlayCount,
					LikeCount:   v.LikeCount,
					Duration:    v.Duration,
				})
			}()

			// 2. 投递 Kafka SummaryTask，由消费者拉起 summary-ai 异步生成摘要
			l.dispatchSummaryTask(v.Id)
		}
	}

	return &video.TranscodeCallbackResp{}, nil
}

// dispatchSummaryTask 投递 AI 摘要任务到 Kafka，并把 video 状态置为「生成中」。
// 完全异步、失败不阻塞主流程；消费端见 internal/consume/summaryconsumer.go。
func (l *TranscodeCallbackLogic) dispatchSummaryTask(videoId int64) {
	writer := l.svcCtx.KafkaSummaryWriter
	if writer == nil {
		l.Logger.Infof("[Summary] KafkaSummaryWriter not configured, skip video_id=%d", videoId)
		return
	}

	// 组装 summary-ai 可直接 GET 的视频 URL（默认 source.mp4 原片，由转码 svc 上传）
	prefix := l.svcCtx.Config.SummaryAI.MinIO
	if prefix == "" {
		prefix = "http://127.0.0.1:9000/" + l.svcCtx.Config.MinIO.Bucket
	}
	task := commonkafka.SummaryTask{
		VideoId:  videoId,
		VideoUrl: fmt.Sprintf("%s/videos/%d/source.mp4", prefix, videoId),
	}
	body, _ := json.Marshal(task)

	// 先把 ai_summary_status 置 1（生成中），便于前端在 status=2 转码完成那一刻就能看到"AI 处理中"
	if err := l.svcCtx.VideoStore.UpdateAiSummaryStatus(l.ctx, videoId, 1); err != nil {
		l.Logger.Errorf("[Summary] mark generating failed: video_id=%d err=%v", videoId, err)
	}

	if err := writer.WriteMessages(l.ctx, kafkago.Message{
		Key:   []byte(fmt.Sprintf("summary-video-%d", videoId)),
		Value: body,
	}); err != nil {
		l.Logger.Errorf("[Summary] kafka write summary task failed: video_id=%d err=%v", videoId, err)
		// 投递失败回滚到 0（未生成），避免前端长时间 loading
		_ = l.svcCtx.VideoStore.UpdateAiSummaryStatus(context.Background(), videoId, 0)
		return
	}
	l.Logger.Infof("[Summary] summary task sent: video_id=%d url=%s", videoId, task.VideoUrl)
}
