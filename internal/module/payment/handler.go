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

// GrantMembership 运营手动发卡（需鉴权；角色校验留 Day6/7）。
func (h *Handler) GrantMembership(c *gin.Context) {
	var req GrantMembershipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	res, err := h.svc.GrantMembership(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, res)
}

// Membership 查询当前用户会员状态（需鉴权，个人中心）。
func (h *Handler) Membership(c *gin.Context) {
	uid := c.GetUint(middleware.ContextUserID)
	res, err := h.svc.MembershipStatus(uid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, res)
}

// ConfirmOrder 模拟支付确认（需鉴权）：确认成功后开通/续期会员。
func (h *Handler) ConfirmOrder(c *gin.Context) {
	orderNo := c.Param("order_no")
	if orderNo == "" {
		response.Error(c, errcode.ErrParam.WithMsg("缺少订单号"))
		return
	}
	uid := c.GetUint(middleware.ContextUserID)
	res, err := h.svc.ConfirmOrder(c.Request.Context(), uid, orderNo)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, res)
}
