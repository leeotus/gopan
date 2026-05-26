package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Elasticsearch struct { Addresses []string; Username, Password string }
}
