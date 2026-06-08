// package ws 提供 WebSocket 弹幕推送 handler。
package ws

import (
	"fmt"
	"net/http"
	"strings"

	"gopan/gateway/internal/svc"

	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// DanmakuHandler 弹幕推送 WebSocket handler。
// 连接参数: ws://host/ws/danmaku?video_id=123&token=xxx
//   - JWT token 验证通过后，订阅 Redis danmaku:{video_id} 频道，实时推送弹幕到客户端。
func DanmakuHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	// 连接一次redis，返回闭包函数，之后每次有用户发送过来ws请求，都会调用这个闭包函数
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", svcCtx.Config.Redis.Host, svcCtx.Config.Redis.Port),
		Password: svcCtx.Config.Redis.Password,
		DB:       svcCtx.Config.Redis.DB,
	})

	// 闭包
	return func(w http.ResponseWriter, r *http.Request) {
		videoId := r.URL.Query().Get("video_id")
		if videoId == "" {
			http.Error(w, "missing video_id", http.StatusBadRequest)
			return
		}

		// JWT 鉴权: 从 query param 或 Authorization header 提取 token
		tokenStr := r.URL.Query().Get("token")
		if tokenStr == "" {
			authHeader := r.Header.Get("Authorization")
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr = parts[1]
			}
		}
		if tokenStr == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(svcCtx.Config.Auth.AccessSecret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("ws upgrade error: %v", err)
			return
		}
		defer conn.Close()

		channel := "danmaku:" + videoId
		// 订阅 Redis 频道，实时推送弹幕到客户端
		pubsub := rdb.Subscribe(r.Context(), channel)
		defer pubsub.Close()

		for msg := range pubsub.Channel() {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				logx.Errorf("ws write error: %v", err)
				break
			}
		}
	}
}
