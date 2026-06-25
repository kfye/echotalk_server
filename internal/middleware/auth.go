package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/echotalk/echotalk_server/internal/pkg/errcode"
	"github.com/echotalk/echotalk_server/internal/pkg/jwt"
	"github.com/echotalk/echotalk_server/internal/pkg/response"
)

// ContextUserID gin 上下文中存放当前用户 ID 的 key。
const ContextUserID = "user_id"

// Auth JWT 鉴权中间件（校验 access 令牌）。
func Auth(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		if token == "" || token == header {
			response.Fail(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}
		claims, err := jwtManager.Parse(token)
		if err != nil || claims.Type != jwt.AccessToken {
			response.Fail(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Set(ContextUserID, claims.UserID)
		c.Next()
	}
}
