// package ws 提供 WebSocket 弹幕推送 handler。
package ws

import (
	"net/http"

	"github.com/redis/go-redis/v9"
	"gopan/gateway/internal/svc"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func DanmakuHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	return func(w http.ResponseWriter, r *http.Request) {
		videoId := r.URL.Query().Get("video_id")
		if videoId == "" {
			http.Error(w, "missing video_id", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("ws upgrade error: %v", err)
			return
		}
		defer conn.Close()

		channel := "danmaku:" + videoId
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
