// package middleware 提供 HTTP 中间件。
// AuthMiddleware 用于保护 /api/video/* 路由，验证 JWT Token。
package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

// AuthMiddleware JWT 鉴权中间件。
// 从 Authorization: Bearer <token> 头中提取 token 并验证。
// 注册到routes.go的middleware中
type AuthMiddleware struct{}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

// Handle 返回 go-zero 标准中间件函数。
// 当前为放行桩实现，TODO: 接入 JWT secret 验证签名并注入 user_id 到 context。
func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK) // 发送200状态码但是显示"未登录"信息
			w.Write([]byte(`{"code":1002,"message":"未登录或登录已过期"}`))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"code":1002,"message":"认证格式错误"}`))
			return
		}

		_ = parts[1] // token string
		_ = jwt.SigningMethodHS256

		// TODO: 验证 JWT 签名，解析 user_id 并注入 context
		// claims := jwt.MapClaims{}
		// jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
		//     return []byte(secret), nil
		// })
		// ctx := context.WithValue(r.Context(), "user_id", claims["user_id"])
		// r = r.WithContext(ctx)

		next(w, r)
	}
}
