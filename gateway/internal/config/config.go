// package config 定义 gateway（API 网关）的配置。
// gateway 作为 HTTP 入口，不直接访问数据库，而是通过 etcd 服务发现调用各个 RPC 服务。
package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config gateway 完整配置，映射 gateway/etc/gateway.yaml。
type Config struct {
	rest.RestConf                // go-zero REST 通用配置（Host/Port 等）
	Auth struct {                // JWT 鉴权配置
		AccessSecret string       // 签名密钥
		AccessExpire int64        // Token 过期时间（秒）
	}
	// 以下为 6 个 RPC 服务的客户端配置，均通过 etcd 服务发现
	UserRpc      zrpc.RpcClientConf // user-svc 的 etcd Key 和 Hosts
	VideoRpc     zrpc.RpcClientConf // video-svc
	TranscodeRpc zrpc.RpcClientConf // transcode-svc
	StreamRpc    zrpc.RpcClientConf // stream-svc
	InteractRpc  zrpc.RpcClientConf // interact-svc
	SearchRpc    zrpc.RpcClientConf // search-svc
}
