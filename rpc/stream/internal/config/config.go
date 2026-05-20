// package config 定义 stream-svc（流媒体服务）的配置。
package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf      // RPC 通用配置
	CDN struct {            // CDN / 对象存储回源配置
		BaseURL   string     // CDN 基础 URL，播放地址的前缀
		SecretKey string     // 防盗链签名密钥
	}
}
