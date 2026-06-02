// package middleware 提供 HTTP 中间件。
// AuthMiddleware 用于保护 /api/video/* 等需要登录的路由，验证 JWT Token。
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

// 注入 context 的 key 类型，避免与第三方包的 context key 冲突。
type ctxKey string

const (
	// CtxKeyUserId 存当前请求用户 ID 的 context key。
	CtxKeyUserId ctxKey = "user_id"
	// CtxKeyUsername 存当前请求用户名的 context key。
	CtxKeyUsername ctxKey = "username"
)

// AuthMiddleware JWT 鉴权中间件。
// 从 Authorization: Bearer <token> 头中提取 token，
// 使用 HMAC-SHA256 验证签名，解析 user_id/username 并注入 request context。
type AuthMiddleware struct {
	secret []byte // HMAC 签名密钥
}

// NewAuthMiddleware 创建 JWT 鉴权中间件。
func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{secret: []byte(secret)}
}

// errorResp 写入统一格式的 JSON 错误响应。
func errorResp(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	body, _ := json.Marshal(map[string]any{
		"code":    code,
		"message": message,
	})
	w.Write(body)
}

// Handle 返回 go-zero 标准中间件函数。
func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 提取 Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			errorResp(w, 1002, "未登录或登录已过期")
			return
		}

		// 2. 校验 Bearer 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			errorResp(w, 1002, "认证格式错误")
			return
		}
		tokenStr := parts[1]

		// 3. 解析并验证 JWT
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			// 校验签名算法必须是 HS256
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return m.secret, nil
		})
		if err != nil || !token.Valid {
			errorResp(w, 1002, "未登录或登录已过期")
			return
		}

		// 4. 提取 claims 并注入 context
		ctx := r.Context()
		if uid, ok := claims["user_id"]; ok {
			uidFloat, ok := uid.(float64)
			if ok {
				ctx = context.WithValue(ctx, CtxKeyUserId, int64(uidFloat))
			}
		}
		if uname, ok := claims["username"]; ok {
			unameStr, ok := uname.(string)
			if ok {
				ctx = context.WithValue(ctx, CtxKeyUsername, unameStr)
			}
		}

		next(w, r.WithContext(ctx))
	}
}

// GetUserIdFromContext 从 context 中提取 JWT 中注入的 user_id。
// 如果 context 中不存在或类型错误，返回 0。
func GetUserIdFromContext(ctx context.Context) int64 {
	uid, ok := ctx.Value(CtxKeyUserId).(int64)
	if !ok {
		return 0
	}
	return uid
}

// GetUsernameFromContext 从 context 中提取 JWT 中注入的 username。
// 如果 context 中不存在或类型错误，返回空字符串。
func GetUsernameFromContext(ctx context.Context) string {
	uname, ok := ctx.Value(CtxKeyUsername).(string)
	if !ok {
		return ""
	}
	return uname
}

// InjectUserIdFromToken 解析 JWT token 并将 user_id/username 注入 context。
// 用于 sendBeacon 等无法携带自定义 header 的场景。
func InjectUserIdFromToken(ctx context.Context, tokenStr, secret string) context.Context {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return ctx
	}
	if uid, ok := claims["user_id"]; ok {
		if uidFloat, ok := uid.(float64); ok {
			ctx = context.WithValue(ctx, CtxKeyUserId, int64(uidFloat))
		}
	}
	if uname, ok := claims["username"]; ok {
		if unameStr, ok := uname.(string); ok {
			ctx = context.WithValue(ctx, CtxKeyUsername, unameStr)
		}
	}
	return ctx
}
