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
	"runtime"
	"strings"
)

// Error 业务错误：携带错误码、提示信息与对应 HTTP 状态。
type Error struct {
	Code int
	Msg  string
	HTTP int

	cause  error  // 被包裹的根因（DB/Redis/讯飞等），仅内部失败时由 Wrap 附带，不返回给客户端
	caller string // 调用 Wrap 的位置 file:line，用于定位是哪一行 err 判断触发的内部失败
}

func (e *Error) Error() string { return fmt.Sprintf("[%d] %s", e.Code, e.Msg) }

// Wrap 返回副本并附带根因与调用点（runtime.Caller(1) 取调用 Wrap 的那一行），用于内部失败定位。
// 拷贝语义，绝不改动包级单例（如 ErrServer）。
func (e *Error) Wrap(cause error) *Error {
	cp := *e
	cp.cause = cause
	if _, file, line, ok := runtime.Caller(1); ok {
		cp.caller = fmt.Sprintf("%s:%d", trimPath(file), line)
	}
	return &cp
}

// Cause 返回被包裹的根因，无则 nil（供日志中间件取根因）。
func (e *Error) Cause() error { return e.cause }

// Caller 返回 Wrap 捕获的调用点 file:line，无则空串（供日志中间件定位）。
func (e *Error) Caller() string { return e.caller }

// Unwrap 暴露根因以支持 errors.Is/As。
func (e *Error) Unwrap() error { return e.cause }

// trimPath 把绝对路径截短到 internal/ 之后，仅为日志可读性；找不到则原样返回。
func trimPath(file string) string {
	if i := strings.Index(file, "internal/"); i >= 0 {
		return file[i:]
	}
	return file
}

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
	// ErrInvalidCredentials 登录失败统一文案，防用户枚举（不区分账号不存在/密码错误）。
	ErrInvalidCredentials = New(11006, "账号或密码错误")
)

// content 12xxx
var (
	ErrContentNotFound = NewHTTP(12001, http.StatusNotFound, "内容不存在")
)

// payment 14xxx
var (
	ErrOrderNotFound   = New(14001, "订单不存在")
	ErrPayFailed       = New(14002, "支付失败")
	ErrProductNotFound = New(14003, "商品不存在")
	ErrProductOffline  = New(14004, "商品已下架")
	ErrOrderStatus     = New(14005, "订单状态异常，无法支付")
)

// speech 15xxx
var (
	ErrSpeechUnavailable = New(15001, "语音服务暂不可用")
	ErrAudioFormat       = NewHTTP(15002, http.StatusBadRequest, "音频格式不符合要求(16K/16bit/单声道WAV)")
)
