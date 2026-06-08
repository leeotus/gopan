package admin

import (
	"net/http"
	"strconv"

	"gopan/gateway/internal/logic/admin"
	"gopan/gateway/internal/middleware"
	"gopan/gateway/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RejectVideoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		videoId, _ := strconv.ParseInt(r.URL.Query().Get("video_id"), 10, 64)
		adminId := middleware.GetUserIdFromContext(r.Context())

		l := admin.NewRejectVideoLogic(r.Context(), svcCtx)
		resp, err := l.RejectVideo(videoId, adminId)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
