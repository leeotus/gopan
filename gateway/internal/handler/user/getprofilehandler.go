// GetProfileHandler 处理 GET /api/user/profile。
package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gopan/gateway/internal/logic/user"
	"gopan/gateway/internal/svc"
)

func GetProfileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := user.NewGetProfileLogic(r.Context(), svcCtx)
		resp, err := l.GetProfile()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
