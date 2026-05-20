// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gopan/gateway/internal/logic/video"
	"gopan/gateway/internal/svc"
)

func DeleteCommentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := video.NewDeleteCommentLogic(r.Context(), svcCtx)
		resp, err := l.DeleteComment()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
