package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"

	"github.com/echotalk/echotalk_server/internal/pkg/response"
)

const headerRequestID = "X-Request-ID"

// RequestID 为每个请求分配 request_id：优先用上游传入的头，否则生成短 ID。
// 写入 gin 上下文（供响应体/日志取用）与响应头。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(headerRequestID)
		if id == "" {
			id = genID()
		}
		c.Set(response.RequestIDKey, id)
		c.Header(headerRequestID, id)
		c.Next()
	}
}

func genID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}
