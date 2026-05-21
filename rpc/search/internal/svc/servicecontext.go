// package svc 定义 search-svc 的依赖注入容器。
package svc

import (
	"context"

	"gopan/common/es"
	"gopan/rpc/search/internal/config"

	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	Config   config.Config
	ESClient *es.Client // Elasticsearch 客户端
}

func NewServiceContext(c config.Config) *ServiceContext {
	esClient, err := es.NewClient(
		c.Elasticsearch.Addresses,
		"gopan_videos", // 索引名
		c.Elasticsearch.Username,
		c.Elasticsearch.Password,
	)
	if err != nil {
		logx.Errorf("es client init failed: %v", err)
	}

	// 确保索引存在
	if esClient != nil {
		if err := esClient.EnsureIndex(context.Background()); err != nil {
			logx.Errorf("es ensure index failed: %v", err)
		}
	}

	return &ServiceContext{
		Config:   c,
		ESClient: esClient,
	}
}
