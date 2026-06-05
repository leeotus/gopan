// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package search

import (
	"net/http"
	"strconv"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gopan/gateway/internal/logic/search"
	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
)

func SearchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		page, _ := strconv.Atoi(q.Get("page"))
		if page <= 0 {
			page = 1
		}
		size, _ := strconv.Atoi(q.Get("size"))
		if size <= 0 {
			size = 20
		}

		req := types.SearchReq{
			Keyword:  q.Get("keyword"),
			Page:     page,
			Size:     size,
			Category: q.Get("category"),
			Sort:     q.Get("sort"),
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
