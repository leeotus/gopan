// user-svc 是 GoPan 平台的用户微服务。
// 提供注册、登录、获取/更新个人信息等 gRPC 接口。
//
// 启动: go run user.go -f etc/user.yaml
// 端口: 8081 (由 yaml 配置)
package main

import (
	"flag"
	"fmt"

	"gopan/rpc/user/internal/config"
	"gopan/rpc/user/internal/server"
	"gopan/rpc/user/internal/svc"
	"gopan/rpc/user/user"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/user.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 初始化依赖（DB 连接 + UserStore）
	ctx := svc.NewServiceContext(c)

	// 创建 gRPC Server 并注册服务实现
	// zrpc.MustNewServer创建一个gRPC服务器(对内给服务间调用)
	// 注意和gateway里的区别，gateway是对外提供服务,使用的是
	// rest.MustNewServer
	// MustNewServer第二个参数是监听函数，当有gRPC调用过来的时候会执行这个函数
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		user.RegisterUserServer(grpcServer, server.NewUserServer(ctx))

		// 开发/测试模式下开启 gRPC reflection（方便 grpcurl 调试）
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting user-svc at %s...\n", c.ListenOn)
	s.Start()
}
