// package config 定义 user-svc（用户服务）的配置。
// go-zero 启动时通过 conf.MustLoad 将 etc/user.yaml 反序列化到此处。
package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf        // 继承 RPC 通用配置（ListenOn / Etcd 等）
	DB struct {               // MySQL 连接配置
		DataSource string      // DSN: user:password@tcp(host:port)/dbname?parseTime=true
	}
	JWT struct {              // JWT 签名配置
		AccessSecret string    // HMAC 密钥，生产环境需更换为复杂随机值
		AccessExpire int64     // Token 有效期（秒）
	}
}
