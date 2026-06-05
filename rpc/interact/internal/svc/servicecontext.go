// package svc 定义 interact-svc 的依赖注入容器。
package svc

import (
	"fmt"

	"gopan/rpc/interact/internal/config"
	"gopan/rpc/interact/store"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config        config.Config
	DB            sqlx.SqlConn
	InteractStore *store.InteractStore
	Redis         *redis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DB.DataSource)
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.CacheRedis.Host, c.CacheRedis.Port),
		Password: c.CacheRedis.Password,
		DB:       c.CacheRedis.DB,
	})
	return &ServiceContext{
		Config:        c,
		DB:            conn,
		InteractStore: store.NewInteractStore(conn),
		Redis:         rdb,
	}
}
