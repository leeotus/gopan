// search-svc 是 GoPan 平台的搜索服务。
// 基于 Elasticsearch 提供视频全文搜索、索引管理和索引删除。
//
// 启动: go run search.go -f etc/search.yaml
// 端口: 8086
package main

import (
	"flag"
	"fmt"

	"gopan/rpc/search/internal/config"
	"gopan/rpc/search/internal/server"
	"gopan/rpc/search/internal/svc"
	"gopan/rpc/search/search"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"github.com/zeromicro/go-zero/core/trace"
	
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/search.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	trace.StartAgent(c.Telemetry)
	defer trace.StopAgent()


	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		search.RegisterSearchServer(grpcServer, server.NewSearchServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting search-svc at %s...\n", c.ListenOn)
	s.Start()
}
