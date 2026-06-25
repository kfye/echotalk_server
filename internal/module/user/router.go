package user

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/echotalk/echotalk_server/internal/middleware"
	"github.com/echotalk/echotalk_server/internal/module/user/codesender"
	"github.com/echotalk/echotalk_server/internal/pkg/jwt"
)

// RegisterRoutes 装配 user 模块路由。
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.Manager, sender codesender.CodeSender, store codesender.CodeStore) {
	repo := NewRepository(db)
	svc := NewService(repo, jwtManager, sender, store)
	h := NewHandler(svc)

	pub := rg.Group("/user")
	{
		pub.POST("/send-code", h.SendCode)
		pub.POST("/register", h.Register)
		pub.POST("/login", h.Login)
		pub.POST("/refresh", h.Refresh)
	}

	auth := rg.Group("/user")
	auth.Use(middleware.Auth(jwtManager))
	{
		auth.GET("/profile", h.Profile)
	}
}
