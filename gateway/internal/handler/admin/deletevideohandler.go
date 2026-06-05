package admin

import (
	"net/http"
	"strconv"

	"gopan/gateway/internal/logic/admin"
	"gopan/gateway/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func DeleteVideoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		videoId, _ := strconv.ParseInt(r.URL.Query().Get("video_id"), 10, 64)
		adminId, _ := strconv.ParseInt(r.URL.Query().Get("admin_id"), 10, 64)

		l := admin.NewDeleteVideoLogic(r.Context(), svcCtx)
		resp, err := l.DeleteVideo(videoId, adminId)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
