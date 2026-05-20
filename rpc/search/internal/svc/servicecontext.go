// package svc 定义 search-svc 的依赖注入容器。
// 当前为桩实现，后续需加入 Elasticsearch 客户端连接。
package svc

import "gopan/rpc/search/internal/config"

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{Config: c}
}
