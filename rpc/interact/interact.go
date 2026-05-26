// interact-svc 是 GoPan 平台的互动服务。
// 提供点赞/收藏/评论/弹幕等社区功能。
//
// 启动: go run interact.go -f etc/interact.yaml
// 端口: 8085
package main

import (
	"flag"
	"fmt"

	"gopan/rpc/interact/internal/config"
	"gopan/rpc/interact/internal/server"
	"gopan/rpc/interact/internal/svc"
	"gopan/rpc/interact/interact"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"github.com/zeromicro/go-zero/core/trace"
	
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/interact.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	trace.StartAgent(c.Telemetry)
	defer trace.StopAgent()


	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		interact.RegisterInteractServer(grpcServer, server.NewInteractServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting interact-svc at %s...\n", c.ListenOn)
	s.Start()
}
