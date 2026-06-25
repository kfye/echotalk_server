package middleware

import "github.com/gin-gonic/gin"

// RateLimit 限流中间件预留：用于语音评测 / AI 调用的并发与成本保护。
// 当前为放行占位，后续可接入令牌桶或 Redis 计数。
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
