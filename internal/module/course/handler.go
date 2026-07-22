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

// --- 管理端：训练营 / 课 / 报名（需鉴权；角色留后续）---

// AdminCourses 训练营分页列表（含草稿），可按 status 过滤。
func (h *Handler) AdminCourses(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	var status *int8
	if s := c.Query("status"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 8); err == nil {
			v := int8(n)
			status = &v
		}
	}
	items, total, page, pageSize, err := h.svc.ListCoursesAdmin(status, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, items, total, page, pageSize)
}

// AdminCreateCourse 新增训练营。
func (h *Handler) AdminCreateCourse(c *gin.Context) {
	var in CourseInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	item, err := h.svc.CreateCourse(in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// AdminUpdateCourse 编辑训练营。
func (h *Handler) AdminUpdateCourse(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	var in CourseInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	item, err := h.svc.UpdateCourse(id, in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// AdminUpdateCourseStatus 训练营上下架。
func (h *Handler) AdminUpdateCourseStatus(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	var in CourseStatusInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	item, err := h.svc.UpdateCourseStatus(id, *in.Status)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// AdminDeleteCourse 软删训练营。
func (h *Handler) AdminDeleteCourse(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if err := h.svc.DeleteCourse(id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// AdminLessons 列某训练营全部课（含词句卡）。
func (h *Handler) AdminLessons(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	items, err := h.svc.ListLessons(id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, items)
}

// AdminCreateLesson 在某训练营下新增每日课。
func (h *Handler) AdminCreateLesson(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	var in LessonInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	item, err := h.svc.CreateLesson(id, in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// AdminUpdateLesson 编辑每日课（词句卡整替换）。
func (h *Handler) AdminUpdateLesson(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	var in LessonInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	item, err := h.svc.UpdateLesson(id, in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// AdminDeleteLesson 软删每日课。
func (h *Handler) AdminDeleteLesson(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if err := h.svc.DeleteLesson(id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// AdminEnroll 手动报名。
func (h *Handler) AdminEnroll(c *gin.Context) {
	var req EnrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	res, err := h.svc.AdminEnroll(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, res)
}

// parseID 从路径取 :id 并转 uint。
func parseID(c *gin.Context) (uint, error) {
	n, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(n), err
}
