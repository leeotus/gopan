// transcode-svc 是 GoPan 平台的转码微服务。
// 负责调用 FFmpeg 将上传的视频源文件转码为多码率 HLS 流。
// 通过 Kafka 消费转码任务，异步处理。
//
// 启动: go run transcode.go -f etc/transcode.yaml
// 端口: 8083
package main

import (
	"context"
	"flag"
	"fmt"

	"gopan/rpc/transcode/internal/config"
	"gopan/rpc/transcode/internal/consume"
	"gopan/rpc/transcode/internal/server"
	"gopan/rpc/transcode/internal/svc"
	"gopan/rpc/transcode/transcode"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"github.com/zeromicro/go-zero/core/trace"
	
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/transcode.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	trace.StartAgent(c.Telemetry)
	defer trace.StopAgent()


	ctx := svc.NewServiceContext(c)

	// 启动 Kafka Consumer（异步消费转码任务）
	go consume.StartConsumer(context.Background(), ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		transcode.RegisterTranscodeServer(grpcServer, server.NewTranscodeServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting transcode-svc at %s...\n", c.ListenOn)
	s.Start()
}
