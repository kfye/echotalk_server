// Package errcode 定义统一错误码与错误类型，配合统一响应结构使用。
//
// 约定（全后端遵循）：
//   - 框架/传输级错误映射真实 HTTP 状态（鉴权401/参数400/无权403/找不到404/限流429/panic500）。
//   - 纯业务拒绝使用 HTTP 200 + 业务码（让客户端当正常响应渲染提示）。
//   - 错误码按模块千号段划分：10xxx通用 / 11xxx user / 12xxx content / 13xxx training /
//     14xxx payment / 15xxx speech / 16xxx ops。
package errcode

import (
	"fmt"
	"net/http"
)

// Error 业务错误：携带错误码、提示信息与对应 HTTP 状态。
type Error struct {
	Code int
	Msg  string
	HTTP int
}

func (e *Error) Error() string { return fmt.Sprintf("[%d] %s", e.Code, e.Msg) }

// New 创建业务类错误码，默认 HTTP 200（业务拒绝）。
func New(code int, msg string) *Error {
	return &Error{Code: code, Msg: msg, HTTP: http.StatusOK}
}

// NewHTTP 创建带显式 HTTP 状态的错误码（框架/传输级）。
func NewHTTP(code, httpStatus int, msg string) *Error {
	return &Error{Code: code, Msg: msg, HTTP: httpStatus}
}

// WithMsg 返回副本并替换提示，不改动码与 HTTP 状态（用于附带校验详情等）。
func (e *Error) WithMsg(msg string) *Error {
	cp := *e
	cp.Msg = msg
	return &cp
}

// 通用 10xxx —— 语义 HTTP 状态
var (
	ErrServer          = NewHTTP(10500, http.StatusInternalServerError, "服务器内部错误")
	ErrParam           = NewHTTP(10400, http.StatusBadRequest, "参数错误")
	ErrUnauthorized    = NewHTTP(10401, http.StatusUnauthorized, "未授权或登录失效")
	ErrForbidden       = NewHTTP(10403, http.StatusForbidden, "无访问权限")
	ErrNotFound        = NewHTTP(10404, http.StatusNotFound, "资源不存在")
	ErrTooManyRequests = NewHTTP(10429, http.StatusTooManyRequests, "请求过于频繁")
)

// user 11xxx —— 业务 200（鉴权类除外）
var (
	ErrUserNotFound  = New(11001, "用户不存在")
	ErrUserExists    = New(11002, "用户已存在")
	ErrPasswordWrong = New(11003, "密码错误")
	ErrTokenInvalid  = NewHTTP(11004, http.StatusUnauthorized, "令牌无效")
	ErrCodeInvalid   = New(11005, "验证码错误或已过期")
)

// payment 14xxx
var (
	ErrOrderNotFound = New(14001, "订单不存在")
	ErrPayFailed     = New(14002, "支付失败")
)

// speech 15xxx
var (
	ErrSpeechUnavailable = New(15001, "语音服务暂不可用")
	ErrAudioFormat       = NewHTTP(15002, http.StatusBadRequest, "音频格式不符合要求(16K/16bit/单声道WAV)")
)
