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
