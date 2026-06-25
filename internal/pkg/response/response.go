// Package response 提供统一响应结构。所有接口返回 {code, msg, data, request_id}。
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/echotalk/echotalk_server/internal/pkg/errcode"
)

// RequestIDKey gin 上下文中存放 request_id 的 key。
// 定义在此（而非 middleware）以避免 response ↔ middleware 循环依赖。
const RequestIDKey = "request_id"

// Body 统一响应体。
type Body struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// PageData 分页数据约定，所有列表接口统一使用。
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func reqID(c *gin.Context) string { return c.GetString(RequestIDKey) }

// Success 成功响应。
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: 0, Msg: "ok", Data: data, RequestID: reqID(c)})
}

// SuccessPage 分页成功响应。
func SuccessPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	Success(c, PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

// Fail 以业务错误码失败响应，HTTP 状态取自 e.HTTP。
func Fail(c *gin.Context, e *errcode.Error) {
	c.JSON(e.HTTP, Body{Code: e.Code, Msg: e.Msg, RequestID: reqID(c)})
}

// Error 统一错误出口：*errcode.Error 按其 HTTP 状态返回；其它 error 归一为 ErrServer(500)，不向客户端暴露内部细节。
func Error(c *gin.Context, err error) {
	if e, ok := err.(*errcode.Error); ok {
		Fail(c, e)
		return
	}
	Fail(c, errcode.ErrServer)
}
