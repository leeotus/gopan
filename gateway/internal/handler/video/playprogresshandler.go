package video

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gopan/gateway/internal/logic/video"
	"gopan/gateway/internal/svc"
)

func SavePlayProgressHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct {
			VideoId  int64   `json:"video_id"`
			UserId   int64   `json:"user_id"`
			Position float64 `json:"position"`
		}
		json.Unmarshal(body, &req)

		l := video.NewSavePlayProgressLogic(r.Context(), svcCtx)
		resp, err := l.SavePlayProgress(req.VideoId, req.UserId, req.Position)
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
		userIdStr := r.URL.Query().Get("user_id")

		l := video.NewGetPlayProgressLogic(r.Context(), svcCtx)
		resp, err := l.GetPlayProgress(videoIdStr, userIdStr)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
