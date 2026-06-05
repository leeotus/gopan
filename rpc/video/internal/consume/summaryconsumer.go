// package consume 在 video-svc 内运行的 Kafka 消费者。
// 目前仅一个消费者：SummaryConsumer，拉取 gopan.video.summary.tasks 触发 summary-ai 听译+摘要。
package consume

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gopan/common/kafka"
	"gopan/rpc/video/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// summaryAIResponse 与 summary-ai /analyze 返回结构对齐。
type summaryAIResponse struct {
	Summary  string `json:"summary"`
	FullText string `json:"full_text"`
	Vtt      string `json:"vtt"`
	Srt      string `json:"srt"`
}

// StartSummaryConsumer 阻塞运行 AI 摘要消费者。
// 通常作为 goroutine 在 video-svc main 中启动。
func StartSummaryConsumer(ctx context.Context, svcCtx *svc.ServiceContext) {
	topic := svcCtx.Config.Kafka.SummaryTopic
	if topic == "" {
		logx.Info("[SummaryConsumer] SummaryTopic empty, consumer disabled")
		return
	}

	reader := kafka.NewConsumer(svcCtx.Config.Kafka.Brokers, topic, "gopan-summary-worker")
	defer reader.Close()

	logx.Infof("[SummaryConsumer] started, topic=%s brokers=%v", topic, svcCtx.Config.Kafka.Brokers)

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logx.Errorf("[SummaryConsumer] fetch error: %v", err)
			continue
		}

		var task kafka.SummaryTask
		if err := json.Unmarshal(msg.Value, &task); err != nil {
			logx.Errorf("[SummaryConsumer] unmarshal error: %v, raw=%s", err, string(msg.Value))
			_ = reader.CommitMessages(ctx, msg)
			continue
		}

		logx.Infof("[SummaryConsumer] received task: video_id=%d url=%s", task.VideoId, task.VideoUrl)
		if err := process(ctx, svcCtx, &task); err != nil {
			logx.Errorf("[SummaryConsumer] process failed: video_id=%d err=%v", task.VideoId, err)
			// 标记失败状态，前端可看到失败提示并触发手动重试
			_ = svcCtx.VideoStore.UpdateAiSummaryStatus(context.Background(), task.VideoId, 3)
		}

		// 成功失败都 commit，防止毒消息死循环；失败重试由前端/管理端显式触发。
		if err := reader.CommitMessages(ctx, msg); err != nil {
			logx.Errorf("[SummaryConsumer] commit error: video_id=%d err=%v", task.VideoId, err)
		}
	}
}

// process 调用 summary-ai HTTP /analyze，并把摘要写回数据库。
func process(ctx context.Context, svcCtx *svc.ServiceContext, task *kafka.SummaryTask) error {
	url := svcCtx.Config.SummaryAI.URL
	if url == "" {
		return fmt.Errorf("SummaryAI.URL not configured")
	}

	timeoutSec := svcCtx.Config.SummaryAI.Timeout
	if timeoutSec <= 0 {
		timeoutSec = 300 // 默认 5 分钟，Whisper tiny 视频几分钟内一般能完成
	}

	// 1. 确保状态为「生成中」（防止生产端未置位的兜底）
	_ = svcCtx.VideoStore.UpdateAiSummaryStatus(ctx, task.VideoId, 1)

	// 2. 发起 HTTP 调用 summary-ai
	payload, _ := json.Marshal(map[string]string{"video_url": task.VideoUrl})
	httpCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(httpCtx, http.MethodPost, url+"/analyze", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build http req: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("call summary-ai: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("summary-ai returned %d: %s", resp.StatusCode, string(raw))
	}

	var aiResp summaryAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return fmt.Errorf("decode summary-ai resp: %w", err)
	}

	if aiResp.Summary == "" {
		// 服务返回空摘要也算失败
		return fmt.Errorf("summary-ai returned empty summary")
	}

	// 3. 写回 DB，并把状态推进到 2（已完成）
	if err := svcCtx.VideoStore.UpdateAiSummary(ctx, task.VideoId, aiResp.Summary); err != nil {
		return fmt.Errorf("save summary: %w", err)
	}

	logx.Infof("[SummaryConsumer] done: video_id=%d elapsed=%s len=%d",
		task.VideoId, time.Since(start).Round(time.Millisecond), len(aiResp.Summary))
	return nil
}
