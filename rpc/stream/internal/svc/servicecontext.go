// package svc 定义 stream-svc 的依赖注入容器。
// 当前只持有配置，后续可加入 Redis 连接用于播放计数。
package svc

import "gopan/rpc/stream/internal/config"

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{Config: c}
}
