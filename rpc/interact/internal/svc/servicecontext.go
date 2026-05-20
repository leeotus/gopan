// package svc 定义 interact-svc 的依赖注入容器。
package svc

import (
	"gopan/rpc/interact/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config
	DB     sqlx.SqlConn // MySQL 连接，用于点赞/收藏/评论/弹幕的持久化
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DB.DataSource)
	return &ServiceContext{
		Config: c,
		DB:     conn,
	}
}
