package course

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/echotalk/echotalk_server/internal/middleware"
	"github.com/echotalk/echotalk_server/internal/pkg/errcode"
	"github.com/echotalk/echotalk_server/internal/pkg/response"
	"github.com/echotalk/echotalk_server/internal/pkg/validatorx"
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

// LessonDetail 某天课详情（需鉴权；未报名/未解锁按业务码拒绝）。
func (h *Handler) LessonDetail(c *gin.Context) {
	day, err := strconv.Atoi(c.Param("day"))
	if err != nil || day < 1 {
		response.Error(c, errcode.ErrParam.WithMsg("day 非法"))
		return
	}
	uid := c.GetUint(middleware.ContextUserID)
	res, err := h.svc.LessonDetail(uid, day)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, res)
}

// Checkin 打卡（需鉴权；客户端完成核心步骤后调用）。
func (h *Handler) Checkin(c *gin.Context) {
	var req CheckinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	uid := c.GetUint(middleware.ContextUserID)
	res, err := h.svc.Checkin(uid, req.Day)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, res)
}
