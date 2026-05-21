// LoginHandler 处理 POST /api/user/login。
package user

import (
	"net/http"

	"gopan/gateway/internal/logic/user"
	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 用于routes.go里注册放行函数
func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	// 使用闭包写法，将svcCtx这个上下文对象传入到放行函数中去
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
