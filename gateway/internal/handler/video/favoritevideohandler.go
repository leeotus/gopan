// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gopan/gateway/internal/logic/video"
	"gopan/gateway/internal/svc"
)

func FavoriteVideoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := video.NewFavoriteVideoLogic(r.Context(), svcCtx)
		resp, err := l.FavoriteVideo()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
