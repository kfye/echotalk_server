package content

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/echotalk/echotalk_server/internal/middleware"
	"github.com/echotalk/echotalk_server/internal/pkg/errcode"
	"github.com/echotalk/echotalk_server/internal/pkg/response"
	"github.com/echotalk/echotalk_server/internal/pkg/validatorx"
)

// Handler 内容 HTTP 处理器。
type Handler struct {
	svc *Service
}

// NewHandler 创建处理器。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// --- App ---

// List 内容列表（可选鉴权）。ß
func (h *Handler) List(c *gin.Context) {
	var q ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	uid := c.GetUint(middleware.ContextUserID)
	items, total, page, pageSize, err := h.svc.List(c.Request.Context(), uid, q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, items, total, page, pageSize)
}

// Detail 内容详情（可选鉴权）。
func (h *Handler) Detail(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	uid := c.GetUint(middleware.ContextUserID)
	d, err := h.svc.Detail(c.Request.Context(), uid, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, d)
}

// --- 管理端 ---

// AdminList 管理列表（含草稿/下架）。
func (h *Handler) AdminList(c *gin.Context) {
	var q ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	var status *int8
	if s := c.Query("status"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 8); err == nil {
			v := int8(n)
			status = &v
		}
	}
	items, total, page, pageSize, err := h.svc.AdminList(q, status)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, items, total, page, pageSize)
}

// Create 新增视频。
func (h *Handler) Create(c *gin.Context) {
	var in VideoInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	d, err := h.svc.Create(in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, d)
}

// Update 编辑视频。
func (h *Handler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	var in VideoInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	d, err := h.svc.Update(id, in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, d)
}

// UpdateStatus 上下架 / 免费付费切换。
func (h *Handler) UpdateStatus(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	var in UpdateStatusInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	d, err := h.svc.UpdateStatus(id, in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, d)
}

// Delete 删除视频（软删）。
func (h *Handler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if err := h.svc.Delete(id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

func parseID(c *gin.Context) (uint, error) {
	n, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(n), err
}
