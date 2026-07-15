package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/echotalk/echotalk_server/internal/pkg/errcode"
	"github.com/echotalk/echotalk_server/internal/pkg/response"
)

// AccessLog 请求访问日志中间件。除请求流水外，对带根因的内部失败集中出一条 Error 日志（含调用点+根因）。
func AccessLog(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		reqID := c.GetString(response.RequestIDKey)
		logger.Info("request",
			zap.String("request_id", reqID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)

		// 集中出错误日志：仅对携带根因（或 5xx）的失败记，裸业务/参数码不刷屏。
		// 按严重度分级：5xx 内部失败 → Error；带根因的 4xx(如鉴权 401/403) → Warn。
		for _, ge := range c.Errors {
			e, ok := ge.Err.(*errcode.Error)
			if !ok || (e.Cause() == nil && e.HTTP < 500) {
				continue
			}
			logFn := logger.Warn
			if e.HTTP >= 500 {
				logFn = logger.Error
			}
			logFn("handler error",
				zap.String("request_id", reqID),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("code", e.Code),
				zap.String("caller", e.Caller()),
				zap.Error(e.Cause()),
			)
		}
	}
}
