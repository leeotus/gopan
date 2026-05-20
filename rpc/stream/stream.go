// stream-svc 是 GoPan 平台的流媒体服务。
// 提供播放地址签发（防盗链签名）和播放计数功能。
//
// 启动: go run stream.go -f etc/stream.yaml
// 端口: 8084
package main

import (
	"flag"
	"fmt"

	"gopan/rpc/stream/internal/config"
	"gopan/rpc/stream/internal/server"
	"gopan/rpc/stream/internal/svc"
	"gopan/rpc/stream/stream"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/stream.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		stream.RegisterStreamServer(grpcServer, server.NewStreamServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting stream-svc at %s...\n", c.ListenOn)
	s.Start()
}
