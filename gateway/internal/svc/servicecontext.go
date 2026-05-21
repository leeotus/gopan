// package svc 定义 gateway 的 ServiceContext——go-zero 的依赖注入容器。
// 启动时创建一次，包含所有 RPC 客户端和中间件，注入到每个 handler 和 logic。
package svc

import (
	"gopan/gateway/internal/config"
	"gopan/gateway/internal/middleware"
	"gopan/rpc/interact/interactclient"
	"gopan/rpc/search/searchclient"
	"gopan/rpc/stream/streamclient"
	"gopan/rpc/transcode/transcodeclient"
	"gopan/rpc/user/userclient"
	"gopan/rpc/video/videoclient"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config          config.Config             // 完整配置
	Auth            rest.Middleware           // JWT 鉴权中间件
	UserClient      userclient.User           // user-svc gRPC 客户端
	VideoClient     videoclient.Video         // video-svc gRPC 客户端
	TranscodeClient transcodeclient.Transcode // transcode-svc gRPC 客户端
	StreamClient    streamclient.Stream       // stream-svc gRPC 客户端
	InteractClient  interactclient.Interact   // interact-svc gRPC 客户端
	SearchClient    searchclient.Search       // search-svc gRPC 客户端
}

// NewServiceContext 初始化所有依赖。
// zrpc.MustNewClient 通过 etcd 自动发现目标服务地址。
func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		Auth:   middleware.NewAuthMiddleware().Handle, // 鉴权函数
		/**
		 * @note zrpc.MustNewClient建立gRPC连接
		 * xxxclient.NewXXX创建一个XXXClient的包装器
		 */
		UserClient:      userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		VideoClient:     videoclient.NewVideo(zrpc.MustNewClient(c.VideoRpc)),
		TranscodeClient: transcodeclient.NewTranscode(zrpc.MustNewClient(c.TranscodeRpc)),
		StreamClient:    streamclient.NewStream(zrpc.MustNewClient(c.StreamRpc)),
		InteractClient:  interactclient.NewInteract(zrpc.MustNewClient(c.InteractRpc)),
		SearchClient:    searchclient.NewSearch(zrpc.MustNewClient(c.SearchRpc)),
	}
}

/**
 * @note 网关的ServiceContext
 * gateway的ServiceContext 持有所有gRPC Client + Auth中间件 + 对应的Config配置
 * 负责: 鉴权 + 转发请求到下游的RPC服务
 * 类比: 路由器 + 电话总机
 *
 * 而在其他地方的ServiceContext, 如user-svc的ServiceContext,
 * 则持有DB连接 + UserStore + 对应的Config配置
 * 负责: 处理具体的业务逻辑(直接操作数据库)
 * 类比: 收银台 + 账本
 */
