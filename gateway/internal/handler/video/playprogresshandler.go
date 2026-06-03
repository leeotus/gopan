package video

import (
	"encoding/json"
	"io"
	"net/http"

	"gopan/gateway/internal/middleware"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gopan/gateway/internal/logic/video"
	"gopan/gateway/internal/svc"
)

func SavePlayProgressHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct {
			VideoId  int64   `json:"video_id"`
			Position float64 `json:"position"`
			Token    string  `json:"token"`
		}
		json.Unmarshal(body, &req)

		// sendBeacon 不支持自定义 header，从 body 中提取 token 注入 context
		if req.Token != "" {
			ctx := middleware.InjectUserIdFromToken(r.Context(), req.Token, svcCtx.Config.Auth.AccessSecret)
			r = r.WithContext(ctx)
		}

		l := video.NewSavePlayProgressLogic(r.Context(), svcCtx)
		resp, err := l.SavePlayProgress(req.VideoId, req.Position)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func GetPlayProgressHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		videoIdStr := r.URL.Query().Get("video_id")

		l := video.NewGetPlayProgressLogic(r.Context(), svcCtx)
		resp, err := l.GetPlayProgress(videoIdStr)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
