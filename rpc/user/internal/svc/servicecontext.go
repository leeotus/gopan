// package svc 定义 user-svc 的依赖注入容器。
// 持有 DB 连接和 UserStore，供所有 logic 使用。
package svc

import (
	"gopan/rpc/user/internal/config"
	"gopan/rpc/user/store"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config    config.Config     // 配置（含 DB 连接串和 JWT 密钥）
	UserStore *store.UserStore  // users 表的数据访问层
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DB.DataSource) // 创建 MySQL 连接
	return &ServiceContext{
		Config:    c,
		UserStore: store.NewUserStore(conn),
	}
}
