package payment

import (
	"github.com/gin-gonic/gin"

	"github.com/echotalk/echotalk_server/internal/middleware"
	"github.com/echotalk/echotalk_server/internal/pkg/errcode"
	"github.com/echotalk/echotalk_server/internal/pkg/response"
	"github.com/echotalk/echotalk_server/internal/pkg/validatorx"
)

// Handler 支付 HTTP 处理器。
type Handler struct {
	svc *Service
}

// NewHandler 创建处理器。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Products 上架商品列表（付费墙，需鉴权）。
func (h *Handler) Products(c *gin.Context) {
	items, err := h.svc.ListProducts(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, items)
}

// CreateOrder 创建订单（需鉴权）。
func (h *Handler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	uid := c.GetUint(middleware.ContextUserID)
	res, err := h.svc.CreateOrder(c.Request.Context(), uid, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, res)
}
