/*
Kafka 削峰延迟对比测试

场景：video-svc 上传完成 → 同步调转码 vs Kafka 异步调转码

用法：直接 go run
输出：同步调用延迟 vs Kafka 异步延迟对比表

前提：Kafka topic gopan.transcode.tasks 已创建，video-svc / transcode-svc 在运行
*/
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	brokers = "127.0.0.1:9092"
	topic   = "gopan.transcode.tasks"
)

type transcodeTask struct {
	VideoId   int64  `json:"video_id"`
	ObjectKey string `json:"object_key"`
}

func main() {
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║   Kafka 削峰延迟对比测试                ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()

	ctx := context.Background()

	// 1. 同步调用模拟（直接写入 Kafka 并等待 consumer 确认）
	// 注意：无法实现真正的"等待确认"，因为 consumer 是异步的
	// 这里只测 Producer 端延迟

	writer := &kafka.Writer{
		Addr:     kafka.TCP(strings.Split(brokers, ",")...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	fmt.Println("▶ 测试 Producer 写入延迟")

	var totalTime time.Duration
	iterations := 500

	for i := int64(1); i <= int64(iterations); i++ {
		task := transcodeTask{VideoId: i, ObjectKey: fmt.Sprintf("videos/%d/source.mp4", i)}
		body, _ := json.Marshal(task)

		start := time.Now()
		err := writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(fmt.Sprintf("video-%d", i)),
			Value: body,
		})
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("  ✗ write error at %d: %v\n", i, err)
			continue
		}
		totalTime += elapsed
	}

	avg := totalTime / time.Duration(iterations)
	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("Producer 写入 %d 条消息:\n", iterations)
	fmt.Printf("  总耗时:   %v\n", totalTime.Round(time.Millisecond))
	fmt.Printf("  平均延迟: %v\n", avg.Round(time.Microsecond))
	fmt.Printf("  QPS:      %.0f msg/s\n", float64(iterations)/totalTime.Seconds())
	fmt.Println()

	// 2. 对比：假设同步调用转码的延迟（即使转码需要 30s，但 Producer 只需发消息）
	fmt.Println("▶ 削峰效果分析")
	fmt.Println(strings.Repeat("─", 50))

	// 假设 1h 视频转码需要 30s
	transcodeTime := 30 * time.Second
	fmt.Printf("假设 1 个视频转码耗时: %v\n", transcodeTime)
	fmt.Printf("同步调用：用户需要等待 %v 才能返回\n", transcodeTime)
	fmt.Printf("Kafka 异步：用户只需等待 %v（Producer 写入延迟）\n", avg.Round(time.Microsecond))
	fmt.Printf("削峰效果：延迟降低 %.0f 倍\n\n", float64(transcodeTime)/float64(avg))

	fmt.Println("▶ 吞吐量对比（每秒能处理多少上传）")
	syncQPS := 1.0 / transcodeTime.Seconds()
	asyncQPS := 1.0 / avg.Seconds()
	fmt.Printf("同步转码: %.3f 上传/秒\n", syncQPS)
	fmt.Printf("Kafka 异步: %.0f 上传/秒\n", asyncQPS)
	fmt.Printf("吞吐提升: %.0f 倍\n\n", asyncQPS/syncQPS)

	// 3. Prometheus 查询建议
	fmt.Println("▶ 建议的 Prometheus 监控指标")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("  Kafka Producer 延迟: rate(kafka_writer_write_time_ms_sum[1m]) / rate(kafka_writer_write_time_ms_count[1m])\n")
	fmt.Printf("  Kafka Consumer 延迟: rate(kafka_reader_fetch_time_ms_sum[1m]) / rate(kafka_reader_fetch_time_ms_count[1m])\n")
	fmt.Printf("  HTTP 接口 P99:       histogram_quantile(0.99, rate(http_server_requests_duration_ms_bucket[1m]))\n")
}
