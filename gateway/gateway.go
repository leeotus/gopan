// gateway 是 GoPan 平台的 API 网关入口。
package main

import (
	"flag"
	"fmt"
	"net/http"

	"gopan/gateway/internal/config"
	"gopan/gateway/internal/handler"
	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/ws"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/trace"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/gateway.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 启动 OpenTelemetry 分布式追踪
	trace.StartAgent(c.Telemetry)
	defer trace.StopAgent()

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// CORS 中间件（必须最先注册，允许前端直连 8888 进行 multipart 上传）
	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next(w, r)
		}
	})

	ctx := svc.NewServiceContext(c)

	// 全局兜底限流 —— 所有 /api/* 请求的统一保护（500/s, burst 1000）
	server.Use(ctx.RateLimiter)

	// 注册 REST 路由
	handler.RegisterHandlers(server, ctx)

	// 注册 WebSocket 路由
	server.AddRoute(rest.Route{
		Method:  "GET",
		Path:    "/ws/danmaku",
		Handler: ws.DanmakuHandler(ctx),
	})

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
