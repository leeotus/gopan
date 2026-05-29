package video

import (
	"net/http"
	"strconv"

	"gopan/gateway/internal/logic/video"
	"gopan/gateway/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetDanmakusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		videoId, _ := strconv.ParseInt(r.URL.Query().Get("video_id"), 10, 64)
		timeVal, _ := strconv.ParseFloat(r.URL.Query().Get("time"), 64)

		l := video.NewGetDanmakusLogic(r.Context(), svcCtx)
		resp, err := l.GetDanmakus(videoId, timeVal)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
