// package response 提供统一的 API 响应格式和错误码定义。
// 所有 HTTP 接口统一返回 { "code": 0, "message": "success", "data": ... }，
// 业务错误通过非零 code 区分，HTTP 状态码保持 200。
package response

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 业务错误码，统一在网关层映射为 JSON 响应。
const (
	CodeSuccess      = 0    // 成功
	CodeParamError   = 1001 // 请求参数错误
	CodeUnauthorized = 1002 // 未登录或 token 过期
	CodeNotFound     = 1003 // 资源不存在
	CodeInternal     = 1004 // 服务器内部错误
	CodeDuplicate    = 1005 // 资源重复（如用户名已存在）
	CodeForbidden    = 1006 // 无权限访问
)

// codeMessages 错误码 → 中文消息映射表。
var codeMessages = map[int]string{
	CodeSuccess:      "success",
	CodeParamError:   "参数错误",
	CodeUnauthorized: "未登录或登录已过期",
	CodeNotFound:     "资源不存在",
	CodeInternal:     "服务器内部错误",
	CodeDuplicate:    "资源已存在",
	CodeForbidden:    "无权限访问",
}

// Body 是统一的 HTTP 响应体结构。
// Code=0 表示成功，非零表示业务错误；
// Message 是人类可读的描述；
// Data 为可选的有效载荷，成功时携带业务数据。
type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Success 写入成功响应（code=0），HTTP 状态码固定为 200。
func Success(w http.ResponseWriter, r *http.Request, data any) {
	httpx.WriteJson(w, http.StatusOK, &Body{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// Error 写入业务错误响应。
// code 为预定义的错误码，可通过可变参数 msg 覆盖默认消息。
// HTTP 状态码仍为 200，便于前端统一处理。
func Error(w http.ResponseWriter, r *http.Request, code int, msg ...string) {
	message := codeMessages[code]
	if len(msg) > 0 && msg[0] != "" {
		message = msg[0]
	}
	httpx.WriteJson(w, http.StatusOK, &Body{
		Code:    code,
		Message: message,
	})
}
