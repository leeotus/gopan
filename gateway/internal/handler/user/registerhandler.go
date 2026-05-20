// RegisterHandler 处理 POST /api/user/register。
// 解析 JSON → 调用 RegisterLogic → 返回 JSON。
package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gopan/gateway/internal/logic/user"
	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
)

func RegisterHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RegisterReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewRegisterLogic(r.Context(), svcCtx)
		resp, err := l.Register(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
