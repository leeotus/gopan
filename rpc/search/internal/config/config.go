// package config 定义 search-svc（搜索服务）的配置。
package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf           // RPC 通用配置
	Elasticsearch struct {       // Elasticsearch 集群地址
		Addresses []string        // 如 ["http://elasticsearch:9200"]
		Username  string
		Password  string
	}
}
