package payment

import (
	"strconv"

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

// --- 管理端：SKU 管理（需鉴权；角色校验留后续）---

// AdminProducts 管理端商品分页列表（含下架），可按 status 过滤。
func (h *Handler) AdminProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	var status *int8
	if s := c.Query("status"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 8); err == nil {
			v := int8(n)
			status = &v
		}
	}
	items, total, page, pageSize, err := h.svc.AdminListProducts(status, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, items, total, page, pageSize)
}

// AdminOrders 管理端订单分页列表（联表带邮箱/商品名），可按 status、user_id 过滤。
func (h *Handler) AdminOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	var status *int8
	if s := c.Query("status"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 8); err == nil {
			v := int8(n)
			status = &v
		}
	}
	var userID *uint
	if u := c.Query("user_id"); u != "" {
		if n, err := strconv.ParseUint(u, 10, 64); err == nil {
			v := uint(n)
			userID = &v
		}
	}
	items, total, page, pageSize, err := h.svc.AdminListOrders(status, userID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, items, total, page, pageSize)
}

// CreateProduct 新增 SKU。
func (h *Handler) CreateProduct(c *gin.Context) {
	var in ProductInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	item, err := h.svc.CreateProduct(in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// UpdateProduct 编辑 SKU。
func (h *Handler) UpdateProduct(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	var in ProductInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	item, err := h.svc.UpdateProduct(id, in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// UpdateProductStatus 上下架 SKU。
func (h *Handler) UpdateProductStatus(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	var in ProductStatusInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	item, err := h.svc.UpdateProductStatus(id, *in.Status)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// DeleteProduct 软删 SKU。
func (h *Handler) DeleteProduct(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if err := h.svc.DeleteProduct(id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// parseID 从路径取 :id 并转 uint。
func parseID(c *gin.Context) (uint, error) {
	n, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(n), err
}
