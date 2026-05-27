package svc

import (
	"gopan/rpc/admin/internal/config"
	"gopan/rpc/admin/store"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config     config.Config
	AdminStore *store.AdminStore
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DB.DataSource)
	return &ServiceContext{
		Config:     c,
		AdminStore: store.NewAdminStore(conn),
	}
}
