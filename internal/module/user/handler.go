package user

import (
	"github.com/gin-gonic/gin"

	"github.com/echotalk/echotalk_server/internal/middleware"
	"github.com/echotalk/echotalk_server/internal/pkg/errcode"
	"github.com/echotalk/echotalk_server/internal/pkg/response"
	"github.com/echotalk/echotalk_server/internal/pkg/validatorx"
)

// Handler 用户 HTTP 处理器。
type Handler struct {
	svc *Service
}

// NewHandler 创建处理器。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// SendCode 发送注册验证码。
func (h *Handler) SendCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	if err := h.svc.SendCode(c.Request.Context(), req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// Register 注册。
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	u, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, u)
}

// Login 登录。
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	pair, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, pair)
}

// Refresh 刷新令牌。
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	pair, err := h.svc.Refresh(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, pair)
}

// Logout 登出：撤销当前用户全部 refresh 会话（需鉴权）。
func (h *Handler) Logout(c *gin.Context) {
	uid := c.GetUint(middleware.ContextUserID)
	if err := h.svc.Logout(c.Request.Context(), uid); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// Profile 个人中心（需鉴权）。
func (h *Handler) Profile(c *gin.Context) {
	uid := c.GetUint(middleware.ContextUserID)
	u, err := h.svc.Profile(uid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, u)
}
