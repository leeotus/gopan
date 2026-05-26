// package kafka 提供 Kafka Producer / Consumer 统一封装和共享类型。
// video-svc 使用 NewProducer 发送转码任务，
// transcode-svc 使用 NewConsumer 消费并调用 processTranscode。
package kafka

import (
	"github.com/segmentio/kafka-go"
)

// ── 共享消息体 ──

// TranscodeTask 发送/消费的转码任务消息体。
type TranscodeTask struct {
	VideoId   int64  `json:"video_id"`
	ObjectKey string `json:"object_key"`
}

// ── 共享 Topic 常量 ──

const (
	TopicTranscodeTasks = "gopan.transcode.tasks"
)

// ── Producer ──

// NewProducer 创建 Kafka Writer（Producer）。
func NewProducer(brokers []string, topic string) *kafka.Writer {
	if topic == "" {
		topic = TopicTranscodeTasks
	}
	return &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{}, // 按 partition 消息数最少的分配
		Logger:   nil,
	}
}

// ── Consumer ──

// NewConsumer 创建 Kafka Reader（Consumer），不指定 GroupID 以避免 KRaft coordinator 问题。
func NewConsumer(brokers []string, topic string) *kafka.Reader {
	if topic == "" {
		topic = TopicTranscodeTasks
	}
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:   brokers,
		Topic:     topic,
		Partition: 0,
		MinBytes:  1,
		MaxBytes:  10e6,
		Logger:    nil,
	})
}
