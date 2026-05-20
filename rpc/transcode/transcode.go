// transcode-svc 是 GoPan 平台的转码微服务。
// 负责调用 FFmpeg 将上传的视频源文件转码为多码率 HLS 流。
//
// 启动: go run transcode.go -f etc/transcode.yaml
// 端口: 8083
package main

import (
	"flag"
	"fmt"

	"gopan/rpc/transcode/internal/config"
	"gopan/rpc/transcode/internal/server"
	"gopan/rpc/transcode/internal/svc"
	"gopan/rpc/transcode/transcode"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/transcode.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

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
