package training

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/echotalk/echotalk_server/internal/middleware"
	"github.com/echotalk/echotalk_server/internal/pkg/jwt"
	"github.com/echotalk/echotalk_server/internal/speech"
)

// RegisterRoutes 装配 training 模块路由。评测一律经 speech 网关。
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.Manager, gw *speech.Gateway) {
	repo := NewRepository(db)
	svc := NewService(repo, gw)
	h := NewHandler(svc)

	// 训练接口需登录（付费门禁留 Day5 会员后再加）
	g := rg.Group("/training")
	g.Use(middleware.Auth(jwtManager))
	{
		g.POST("/evaluate", h.Evaluate)
		g.GET("/records", h.Records)
	}
}
