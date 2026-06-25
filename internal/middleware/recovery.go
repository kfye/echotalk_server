package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/echotalk/echotalk_server/internal/pkg/errcode"
	"github.com/echotalk/echotalk_server/internal/pkg/response"
)

// Recovery 全局异常中间件：recover panic 并返回统一错误响应，绝不裸抛 500。
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					zap.Any("error", r),
					zap.String("path", c.Request.URL.Path),
					zap.String("request_id", c.GetString(response.RequestIDKey)),
					zap.Stack("stack"),
				)
				response.Error(c, errcode.ErrServer)
				c.Abort()
			}
		}()
		c.Next()
	}
}
