// Package router 装配 gin 引擎、全局中间件，并按模块注册路由。
package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/echotalk/echotalk_server/internal/config"
	"github.com/echotalk/echotalk_server/internal/middleware"
	"github.com/echotalk/echotalk_server/internal/module/content"
	"github.com/echotalk/echotalk_server/internal/module/payment"
	"github.com/echotalk/echotalk_server/internal/module/payment/channel"
	"github.com/echotalk/echotalk_server/internal/module/user"
	"github.com/echotalk/echotalk_server/internal/module/user/codesender"
	"github.com/echotalk/echotalk_server/internal/module/user/tokenstore"
	"github.com/echotalk/echotalk_server/internal/pkg/jwt"
	"github.com/echotalk/echotalk_server/internal/pkg/response"
	"github.com/echotalk/echotalk_server/internal/speech"
)

// Deps 路由装配所需依赖。
type Deps struct {
	Config     *config.Config
	DB         *gorm.DB
	Logger     *zap.Logger
	JWT        *jwt.Manager
	Speech     *speech.Gateway
	PayChannel channel.PaymentChannel
	CodeSender codesender.CodeSender
	CodeStore  codesender.CodeStore
	TokenStore tokenstore.TokenStore
	Members    content.MembershipChecker
}

// Setup 构建 gin 引擎并注册所有模块路由。
func Setup(d Deps) *gin.Engine {
	gin.SetMode(d.Config.Server.Mode)
	engine := gin.New()
	engine.Use(
		middleware.RequestID(),
		middleware.Recovery(d.Logger),
		middleware.AccessLog(d.Logger),
		middleware.CORS(),
	)

	engine.GET("/healthz", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	api := engine.Group("/api/v1")
	user.RegisterRoutes(api, d.DB, d.JWT, d.CodeSender, d.CodeStore, d.TokenStore)
	content.RegisterRoutes(api, d.DB, d.JWT, d.Members)
	payment.RegisterRoutes(api, d.DB, d.JWT, d.PayChannel)
	// training / ops 模块路由待接入。
	// training 模块将通过 d.Speech 调用语音评测网关。
	_ = d.Speech

	return engine
}
