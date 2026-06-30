package payment

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/echotalk/echotalk_server/internal/middleware"
	"github.com/echotalk/echotalk_server/internal/module/payment/channel"
	"github.com/echotalk/echotalk_server/internal/pkg/jwt"
)

// RegisterRoutes 装配 payment 模块路由。
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.Manager, ch channel.PaymentChannel) {
	repo := NewRepository(db)
	svc := NewService(repo, ch)
	h := NewHandler(svc)

	g := rg.Group("/payment")
	g.Use(middleware.Auth(jwtManager))
	{
		g.GET("/products", h.Products)
		g.GET("/membership", h.Membership)
		g.POST("/orders", h.CreateOrder)
		g.POST("/orders/:order_no/confirm", h.ConfirmOrder)
	}

	// 管理端：手动发卡（与 content /admin/* 一致，先只校验登录，角色校验留 Day6/7）
	adminMember := rg.Group("/admin/memberships")
	adminMember.Use(middleware.Auth(jwtManager))
	{
		adminMember.POST("/grant", h.GrantMembership)
	}

	// 管理端：SKU 管理（镜像 content /admin/videos，先只校验登录）
	adminProduct := rg.Group("/admin/products")
	adminProduct.Use(middleware.Auth(jwtManager))
	{
		adminProduct.GET("", h.AdminProducts)
		adminProduct.POST("", h.CreateProduct)
		adminProduct.PUT("/:id", h.UpdateProduct)
		adminProduct.PATCH("/:id/status", h.UpdateProductStatus)
		adminProduct.DELETE("/:id", h.DeleteProduct)
	}

	// 管理端：订单列表（只读，联表带邮箱/商品名）
	adminOrder := rg.Group("/admin/orders")
	adminOrder.Use(middleware.Auth(jwtManager))
	{
		adminOrder.GET("", h.AdminOrders)
	}
}
