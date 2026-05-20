// gateway 是 GoPan 平台的 API 网关入口。
// 作用：接收客户端 HTTP 请求 → JWT 鉴权 → 路由到对应的 RPC 服务 → 返回统一格式 JSON。
//
// 启动: go run gateway.go -f etc/gateway.yaml
// 端口: 8888 (由 yaml 配置)
package main

import (
	"flag"
	"fmt"

	"gopan/gateway/internal/config"
	"gopan/gateway/internal/handler"
	"gopan/gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/gateway.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c) // 将 yaml 反序列化为 Config 结构体

	// 创建 REST server
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 初始化 ServiceContext（依赖注入容器）
	ctx := svc.NewServiceContext(c)

	// 注册所有路由（goctl 自动生成）
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
