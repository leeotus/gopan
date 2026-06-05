// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package admin

import (
	"net/http"
	"strconv"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gopan/gateway/internal/logic/admin"
	"gopan/gateway/internal/svc"
)

func AdminListVideosHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cursor, _ := strconv.ParseInt(r.URL.Query().Get("cursor"), 10, 64)
		status, _ := strconv.ParseInt(r.URL.Query().Get("status"), 10, 32)

		l := admin.NewAdminListVideosLogic(r.Context(), svcCtx)
		resp, err := l.AdminListVideos(cursor, int32(status))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
