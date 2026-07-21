package course

import (
	"github.com/gin-gonic/gin"

	"github.com/echotalk/echotalk_server/internal/middleware"
	"github.com/echotalk/echotalk_server/internal/pkg/response"
)

// Handler 训练营 HTTP 处理器。
type Handler struct {
	svc *Service
}

// NewHandler 创建处理器。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Plans 上架训练营列表（报名墙，需鉴权）。
func (h *Handler) Plans(c *gin.Context) {
	items, err := h.svc.ListPlans(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, items)
}

// My 我的训练营：进度 + 今日课概要（需鉴权）。
func (h *Handler) My(c *gin.Context) {
	uid := c.GetUint(middleware.ContextUserID)
	res, err := h.svc.MyCourse(uid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, res)
}
