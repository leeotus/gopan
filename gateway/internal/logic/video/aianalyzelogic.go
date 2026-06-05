// AIAnalyzeLogic 是 AI 摘要的「查询 + 重试」入口。
//
// 流程：
//   1. 调 video-svc.GetVideo 取最新的 ai_summary_status 与 ai_summary。
//   2. status=2 → 直接返回缓存的摘要。
//   3. status=1 → 返回 generating，前端继续轮询 GetVideo 即可。
//   4. status=0 或 3 → 已不再触发新的 Whisper 计算（避免堵塞 HTTP 入口）。
//      返回当前状态，前端可显示「等待中」或「失败可重试」。
//      重试由后端管理或后续单独接口完成，避免任何用户都能触发昂贵的 AI 任务。
//
// 注意：以前的实现是同步阻塞调 summary-ai（45 秒超时），现已经完全迁移到 Kafka 异步链路。
package video

import (
	"context"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/video/videoclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AIAnalyzeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAIAnalyzeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AIAnalyzeLogic {
	return &AIAnalyzeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AIAnalyzeLogic) AIAnalyze(req *types.AIAnalyzeReq) (*types.AIAnalyzeResp, error) {
	r, err := l.svcCtx.VideoClient.GetVideo(l.ctx, &videoclient.GetVideoReq{VideoId: req.VideoId})
	if err != nil {
		l.Logger.Errorf("[AI Analyze] GetVideo failed: video_id=%d err=%v", req.VideoId, err)
		return nil, err
	}
	v := r.Video

	resp := &types.AIAnalyzeResp{
		Status: int(v.AiSummaryStatus),
	}
	if v.AiSummaryStatus == 2 {
		resp.Summary = v.AiSummary
	}
	return resp, nil
}
