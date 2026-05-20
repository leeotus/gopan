// package svc 定义 gateway 的 ServiceContext——go-zero 的依赖注入容器。
// 启动时创建一次，包含所有 RPC 客户端和中间件，注入到每个 handler 和 logic。
package svc

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"gopan/gateway/internal/config"
	"gopan/gateway/internal/middleware"
	"gopan/rpc/interact/interactclient"
	"gopan/rpc/search/searchclient"
	"gopan/rpc/stream/streamclient"
	"gopan/rpc/transcode/transcodeclient"
	"gopan/rpc/user/userclient"
	"gopan/rpc/video/videoclient"
)

type ServiceContext struct {
	Config         config.Config                  // 完整配置
	Auth           rest.Middleware                // JWT 鉴权中间件
	UserClient     userclient.User               // user-svc gRPC 客户端
	VideoClient    videoclient.Video              // video-svc gRPC 客户端
	TranscodeClient transcodeclient.Transcode     // transcode-svc gRPC 客户端
	StreamClient   streamclient.Stream            // stream-svc gRPC 客户端
	InteractClient interactclient.Interact        // interact-svc gRPC 客户端
	SearchClient   searchclient.Search            // search-svc gRPC 客户端
}

// NewServiceContext 初始化所有依赖。
// zrpc.MustNewClient 通过 etcd 自动发现目标服务地址。
func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:          c,
		Auth:            middleware.NewAuthMiddleware().Handle,
		UserClient:      userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		VideoClient:     videoclient.NewVideo(zrpc.MustNewClient(c.VideoRpc)),
		TranscodeClient: transcodeclient.NewTranscode(zrpc.MustNewClient(c.TranscodeRpc)),
		StreamClient:    streamclient.NewStream(zrpc.MustNewClient(c.StreamRpc)),
		InteractClient:  interactclient.NewInteract(zrpc.MustNewClient(c.InteractRpc)),
		SearchClient:    searchclient.NewSearch(zrpc.MustNewClient(c.SearchRpc)),
	}
}
