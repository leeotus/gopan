// package svc 定义 gateway 的 ServiceContext——go-zero 的依赖注入容器。
package svc

import (
	"fmt"

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
	Config      config.Config
	Auth        rest.Middleware // rest.Middleware => func(next http.HandlerFunc) http.HandlerFunc
	RateLimiter rest.Middleware // 全局兜底限流 (500/s)
	RateLimiterList   rest.Middleware // 视频列表限流 (300/s)
	RateLimiterDetail rest.Middleware // 视频详情限流 (200/s)
	RateLimiterLogin  rest.Middleware // 登录限流 (10/s, 防撞库)

	// 可以将以下的服务客户端看成是java springboot里的@Autowired，go-zero会自动注入
	UserClient      userclient.User
	VideoClient     videoclient.Video
	TranscodeClient transcodeclient.Transcode
	StreamClient    streamclient.Stream
	InteractClient  interactclient.Interact
	SearchClient    searchclient.Search
	AdminClient     adminclient.Admin
}

// tryNewClient 尝试建立 zrpc 客户端，失败不 panic，仅记日志并返回 nil
// TODO: 如果有服务在一开始无法建立连接，希望后续可以自动重试
func tryNewClient(cfg zrpc.RpcClientConf) zrpc.Client {
	cli, err := zrpc.NewClient(cfg)
	if err != nil {
		logx.Errorf("failed to create zrpc client: %v", err)
		return nil
	}
	return cli
}

func NewServiceContext(c config.Config) *ServiceContext {
	// redis客户端对象
	rds := redis.MustNewRedis(redis.RedisConf{
		Host: fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port),
		Type: redis.NodeType,
		Pass: c.Redis.Password,
	})

	ctx := &ServiceContext{
		Config: c,
		// 鉴权中间件
		Auth: middleware.NewAuthMiddleware(c.Auth.AccessSecret).Handle,
		// 分级限流中间件
		//  全局兜底: 所有 /api/* 统一保护
		RateLimiter:       middleware.NewRateLimitMiddleware(rds, 3000, 5000, "gateway:ratelimit:global:").Handle,
		//  视频列表: 首页流量大，宽松
		RateLimiterList:   middleware.NewRateLimitMiddleware(rds, 2000, 3000, "gateway:ratelimit:list:").Handle,
		//  视频详情: 中等流量
		RateLimiterDetail: middleware.NewRateLimitMiddleware(rds, 1000, 2000, "gateway:ratelimit:detail:").Handle,
		//  登录: 低频但敏感，严格防撞库
		RateLimiterLogin:  middleware.NewRateLimitMiddleware(rds, 50, 100, "gateway:ratelimit:login:").Handle,
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
