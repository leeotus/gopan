// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gopan/gateway/internal/logic/video"
	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
)

func InitUploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.InitUploadReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := video.NewInitUploadLogic(r.Context(), svcCtx)
		resp, err := l.InitUpload(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
