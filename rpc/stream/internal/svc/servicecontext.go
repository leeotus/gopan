// package svc 定义 stream-svc 的依赖注入容器。
// 当前只持有配置，后续可加入 Redis 连接用于播放计数。
package svc

import (
	"fmt"

	"gopan/rpc/stream/internal/config"

	"github.com/redis/go-redis/v9"
)

type ServiceContext struct {
	Config config.Config
	Redis  *redis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.CacheRedis.Host, c.CacheRedis.Port),
		Password: c.CacheRedis.Password,
		DB:       c.CacheRedis.DB,
	})
	return &ServiceContext{
		Config: c,
		Redis:  rdb,
	}
}
