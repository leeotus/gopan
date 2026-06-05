// video-svc 是 GoPan 平台的视频微服务。
// 提供视频上传（分片）、元数据管理、列表查询、转码回调等 gRPC 接口。
//
// 启动: go run video.go -f etc/video.yaml
// 端口: 8082
package main

import (
	"context"
	"flag"
	"fmt"

	"gopan/rpc/video/internal/config"
	"gopan/rpc/video/internal/consume"
	"gopan/rpc/video/internal/server"
	"gopan/rpc/video/internal/svc"
	"gopan/rpc/video/video"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/trace"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"

	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/video.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	trace.StartAgent(c.Telemetry)
	defer trace.StopAgent()

	ctx := svc.NewServiceContext(c)

	// 后台启动 AI 摘要 Kafka 消费者（topic 留空时自动 no-op）
	go consume.StartSummaryConsumer(context.Background(), ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		video.RegisterVideoServer(grpcServer, server.NewVideoServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting video-svc at %s...\n", c.ListenOn)
	s.Start()
}
