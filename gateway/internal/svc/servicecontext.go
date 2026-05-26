// package svc 定义 gateway 的 ServiceContext——go-zero 的依赖注入容器。
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

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config          config.Config
	Auth            rest.Middleware
	UserClient      userclient.User
	VideoClient     videoclient.Video
	TranscodeClient transcodeclient.Transcode
	StreamClient    streamclient.Stream
	InteractClient  interactclient.Interact
	SearchClient    searchclient.Search
}

// tryNewClient 尝试建立 zrpc 客户端，失败不 panic，仅记日志并返回 nil
func tryNewClient(cfg zrpc.RpcClientConf) zrpc.Client {
	cli, err := zrpc.NewClient(cfg)
	if err != nil {
		logx.Errorf("failed to create zrpc client: %v", err)
		return nil
	}
	return cli
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:          c,
		Auth:            middleware.NewAuthMiddleware().Handle,
		UserClient:      userclient.NewUser(tryNewClient(c.UserRpc)),
		VideoClient:     videoclient.NewVideo(tryNewClient(c.VideoRpc)),
		TranscodeClient: transcodeclient.NewTranscode(tryNewClient(c.TranscodeRpc)),
		StreamClient:    streamclient.NewStream(tryNewClient(c.StreamRpc)),
		InteractClient:  interactclient.NewInteract(tryNewClient(c.InteractRpc)),
		SearchClient:    searchclient.NewSearch(tryNewClient(c.SearchRpc)),
	}
}
