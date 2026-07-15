package middleware

import (
	"errors"
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
			response.Fail(c, errcode.ErrUnauthorized.Wrap(errors.New("缺少 Bearer token")))
			c.Abort()
			return
		}
		claims, err := jwtManager.Parse(token)
		if err != nil {
			response.Fail(c, errcode.ErrUnauthorized.Wrap(err)) // jwt 解析失败(过期/无效/格式错)
			c.Abort()
			return
		}
		if claims.Type != jwt.AccessToken {
			response.Fail(c, errcode.ErrUnauthorized.Wrap(errors.New("非 access 令牌")))
			c.Abort()
			return
		}
		c.Set(ContextUserID, claims.UserID)
		c.Next()
	}
}

// OptionalAuth 可选鉴权：带有效 access 令牌则注入 userID，无/无效令牌也放行（匿名）。
// 用于内容浏览这类匿名可访问、但登录后能解锁付费内容的接口。
func OptionalAuth(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		if token != "" && token != header {
			if claims, err := jwtManager.Parse(token); err == nil && claims.Type == jwt.AccessToken {
				c.Set(ContextUserID, claims.UserID)
			}
		}
		c.Next()
	}
}
