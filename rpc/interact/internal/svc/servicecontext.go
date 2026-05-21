// package svc 定义 interact-svc 的依赖注入容器。
package svc

import (
	"gopan/rpc/interact/internal/config"
	"gopan/rpc/interact/store"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config        config.Config
	DB            sqlx.SqlConn
	InteractStore *store.InteractStore // 四张互动表的共享数据访问层
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DB.DataSource)
	return &ServiceContext{
		Config:        c,
		DB:            conn,
		InteractStore: store.NewInteractStore(conn),
	}
}
