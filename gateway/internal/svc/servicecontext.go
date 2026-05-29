// package svc 定义 gateway 的 ServiceContext——go-zero 的依赖注入容器。
package svc

import (
	"gopan/gateway/internal/config"
	"gopan/gateway/internal/middleware"
	"gopan/rpc/admin/adminclient"
	"gopan/rpc/interact/interactclient"
	"gopan/rpc/search/searchclient"
	"gopan/rpc/stream/streamclient"
	"gopan/rpc/transcode/transcodeclient"
	"gopan/rpc/user/userclient"
	"gopan/rpc/video/videoclient"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config          config.Config
	Auth            rest.Middleware
	RateLimiter     rest.Middleware
	UserClient      userclient.User
	VideoClient     videoclient.Video
	TranscodeClient transcodeclient.Transcode
	StreamClient    streamclient.Stream
	InteractClient  interactclient.Interact
	SearchClient    searchclient.Search
	AdminClient     adminclient.Admin
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
	rds := redis.MustNewRedis(redis.RedisConf{
		Host: c.Redis.Host,
		Type: redis.NodeType,
		Pass: c.Redis.Password,
	})

	ctx := &ServiceContext{
		Config:      c,
		Auth:        middleware.NewAuthMiddleware(c.Auth.AccessSecret).Handle,
		RateLimiter: middleware.NewRateLimitMiddleware(rds, 100, 200).Handle,
	}

	// 异步初始化 RPC 客户端，避免阻塞启动
	go func() {
		ctx.UserClient = userclient.NewUser(tryNewClient(c.UserRpc))
	}()
	go func() {
		ctx.VideoClient = videoclient.NewVideo(tryNewClient(c.VideoRpc))
	}()
	go func() {
		ctx.TranscodeClient = transcodeclient.NewTranscode(tryNewClient(c.TranscodeRpc))
	}()
	go func() {
		ctx.StreamClient = streamclient.NewStream(tryNewClient(c.StreamRpc))
	}()
	go func() {
		ctx.InteractClient = interactclient.NewInteract(tryNewClient(c.InteractRpc))
	}()
	go func() {
		ctx.SearchClient = searchclient.NewSearch(tryNewClient(c.SearchRpc))
	}()
	go func() {
		ctx.AdminClient = adminclient.NewAdmin(tryNewClient(c.AdminRpc))
	}()

	return ctx
}
