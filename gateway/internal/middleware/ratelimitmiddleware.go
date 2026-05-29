// package middleware 提供 HTTP 中间件。
// RateLimitMiddleware 基于令牌桶实现对公开接口的限流保护。
package middleware

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// RateLimitMiddleware 令牌桶限流中间件。
// 使用 go-zero 内置的 TokenLimiter，基于 Redis + Lua 脚本实现分布式令牌桶。
type RateLimitMiddleware struct {
	limiter *limit.TokenLimiter
}

// NewRateLimitMiddleware 创建限流中间件。
// rate: 每秒生成的令牌数，burst: 桶容量（允许的最大突发）。
func NewRateLimitMiddleware(rds *redis.Redis, rate, burst int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: limit.NewTokenLimiter(rate, burst, rds, "gateway:ratelimit:"),
	}
}

// Handle 返回 go-zero 标准中间件函数。
func (m *RateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !m.limiter.Allow() {
			httpx.WriteJson(w, http.StatusTooManyRequests, map[string]any{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
			})
			return
		}
		next(w, r)
	}
}
