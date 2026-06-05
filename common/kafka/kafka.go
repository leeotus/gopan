// package kafka 提供 Kafka Producer / Consumer 统一封装和共享类型。
// video-svc 使用 NewProducer 发送转码任务，
// transcode-svc 使用 NewConsumer 消费并调用 processTranscode。
package kafka

import (
	kafkago "github.com/segmentio/kafka-go"
)

// ── 共享消息体 ──

// TranscodeTask 发送/消费的转码任务消息体。
type TranscodeTask struct {
	VideoId   int64  `json:"video_id"`
	ObjectKey string `json:"object_key"`
}

// MergeTask 发送/消费的合并任务消息体（用于异步合并）。
type MergeTask struct {
	VideoId     int64    `json:"video_id"`
	UploadId    string   `json:"upload_id"`
	ChunkKeys   []string `json:"chunk_keys"`
	TotalChunks int32    `json:"total_chunks"`
}

// SummaryTask 发送/消费的 AI 摘要任务消息体。
// 转码完成后由 video-svc.TranscodeCallback 投递，summary 消费者拉到后调用 summary-ai 微服务。
type SummaryTask struct {
	VideoId  int64  `json:"video_id"`
	VideoUrl string `json:"video_url"` // summary-ai 可直接 GET 下载的 MinIO/源站 URL
}

// ── 共享 Topic 常量 ──

const (
	TopicTranscodeTasks = "gopan.transcode.tasks"
	TopicMergeTasks     = "gopan.video.merge.tasks"
	TopicSummaryTasks   = "gopan.video.summary.tasks"
)

// ── Producer ──

// NewProducer 创建 Kafka Writer（Producer）。
func NewProducer(brokers []string, topic string) *kafkago.Writer {
	if topic == "" {
		topic = TopicTranscodeTasks
	}
	return &kafkago.Writer{
		Addr:     kafkago.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafkago.LeastBytes{},
		Logger:   nil,
	}
}

// ── Consumer ──

// NewConsumer 创建 Kafka Reader（Consumer），带稳定 GroupID 持久化消费 offset。
// 同一 GroupID 内的多个 reader 自动负载均衡分区；offset commit 后重启不会重读已处理消息。
// 不要把 groupID 留空，否则 offset 仅在内存中且 CommitMessages 会失败。
func NewConsumer(brokers []string, topic, groupID string) *kafkago.Reader {
	if topic == "" {
		topic = TopicTranscodeTasks
	}
	cfg := kafkago.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		MinBytes: 1,
		MaxBytes: 10e6,
		Logger:   nil,
	}
	if groupID != "" {
		// 走 consumer group 模式：分区自动分配 + offset 持久化在 __consumer_offsets。
		cfg.GroupID = groupID
		cfg.StartOffset = kafkago.FirstOffset // 新 group 首次启动从最早消息开始（确保不漏）
	} else {
		// 兼容旧调用：单分区直读，offset 仅在内存。
		cfg.Partition = 0
		cfg.StartOffset = kafkago.LastOffset
	}
	return kafkago.NewReader(cfg)
}
