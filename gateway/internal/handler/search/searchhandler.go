// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package search

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gopan/gateway/internal/logic/search"
	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
)

func SearchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SearchReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := search.NewSearchLogic(r.Context(), svcCtx)
		resp, err := l.Search(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
