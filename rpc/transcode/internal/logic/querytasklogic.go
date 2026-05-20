// QueryTaskLogic 查询转码任务状态。
// TODO: 从 Redis 查询任务进度和结果。
package logic

import (
	"context"

	"gopan/rpc/transcode/internal/svc"
	"gopan/rpc/transcode/transcode"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryTaskLogic {
	return &QueryTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryTaskLogic) QueryTask(in *transcode.QueryTaskReq) (*transcode.QueryTaskResp, error) {
	l.Logger.Infof("query transcode task: taskId=%s, videoId=%d", in.TaskId, in.VideoId)
	return &transcode.QueryTaskResp{Status: 1}, nil // 1 = 转码中
}
