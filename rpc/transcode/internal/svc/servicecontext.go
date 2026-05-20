// package svc 定义 transcode-svc 的依赖注入容器。
// taskSeq 用于生成唯一任务 ID。
package svc

import (
	"sync/atomic"

	"gopan/rpc/transcode/internal/config"
)

type ServiceContext struct {
	Config  config.Config
	taskSeq atomic.Int64 // 自增序列，保证任务 ID 唯一
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{Config: c}
}

// GenTaskId 生成唯一的转码任务 ID。
func (s *ServiceContext) GenTaskId() int64 {
	return s.taskSeq.Add(1)
}
