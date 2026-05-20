// package svc 定义 video-svc 的依赖注入容器。
package svc

import (
	"gopan/rpc/video/internal/config"
	"gopan/rpc/video/store"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config     config.Config
	VideoStore *store.VideoStore // videos 和 transcodes 表的共享数据访问层
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DB.DataSource)
	return &ServiceContext{
		Config:     c,
		VideoStore: store.NewVideoStore(conn),
	}
}
