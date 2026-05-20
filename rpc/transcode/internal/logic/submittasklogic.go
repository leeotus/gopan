// SubmitTaskLogic 提交转码任务。
// 流程: 接收源文件 MinIO key → 异步调用 FFmpeg → HLS 多码率切片 → 回调 video-svc。
package logic

import (
	"context"
	"fmt"

	"gopan/rpc/transcode/internal/svc"
	"gopan/rpc/transcode/transcode"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitTaskLogic {
	return &SubmitTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SubmitTask 提交转码任务。
// 生成唯一任务 ID，后续由 FFmpeg 进程异步处理。
// TODO: 集成 FFmpeg 和 MinIO 客户端完成实际转码流程。
func (l *SubmitTaskLogic) SubmitTask(in *transcode.SubmitTaskReq) (*transcode.SubmitTaskResp, error) {
	taskId := fmt.Sprintf("transcode_%d_%d", in.VideoId, l.svcCtx.GenTaskId())

	l.Logger.Infof("submit transcode task: taskId=%s, videoId=%d, resolutions=%v",
		taskId, in.VideoId, in.Resolutions)

	return &transcode.SubmitTaskResp{TaskId: taskId}, nil
}
