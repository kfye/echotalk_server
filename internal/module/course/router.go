package course

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/echotalk/echotalk_server/internal/middleware"
	"github.com/echotalk/echotalk_server/internal/pkg/jwt"
)

// RegisterRoutes 装配 course 模块路由。
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.Manager) {
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)

	// App 端：训练营（需鉴权）
	g := rg.Group("/course")
	g.Use(middleware.Auth(jwtManager))
	{
		g.GET("/plans", h.Plans)
		g.GET("/my", h.My)
	}
}
