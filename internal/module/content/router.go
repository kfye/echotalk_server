package content

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/echotalk/echotalk_server/internal/middleware"
	"github.com/echotalk/echotalk_server/internal/pkg/jwt"
)

// RegisterRoutes 装配 content 模块路由。
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.Manager, members MembershipChecker) {
	repo := NewRepository(db)
	svc := NewService(repo, members)
	h := NewHandler(svc)

	// App 端：可选鉴权（匿名可浏览，登录后解锁付费）
	app := rg.Group("/content/videos")
	app.Use(middleware.OptionalAuth(jwtManager))
	{
		app.GET("", h.List)
		app.GET("/:id", h.Detail)
	}

	// 管理端：需鉴权（运营角色校验待 Day6/7 补，先只校验登录）
	admin := rg.Group("/admin/videos")
	admin.Use(middleware.Auth(jwtManager))
	{
		admin.GET("", h.AdminList)
		admin.POST("", h.Create)
		admin.PUT("/:id", h.Update)
		admin.PATCH("/:id/status", h.UpdateStatus)
		admin.DELETE("/:id", h.Delete)
	}
}
