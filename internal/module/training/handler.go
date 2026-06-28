package training

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/echotalk/echotalk_server/internal/middleware"
	"github.com/echotalk/echotalk_server/internal/pkg/errcode"
	"github.com/echotalk/echotalk_server/internal/pkg/response"
	"github.com/echotalk/echotalk_server/internal/pkg/validatorx"
)

// maxAudioBytes 上传录音大小上限(8MB)，防小内存机器 OOM。
const maxAudioBytes = 8 << 20

// Handler 训练 HTTP 处理器。
type Handler struct {
	svc *Service
}

// NewHandler 创建处理器。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Evaluate 上传录音评测（鉴权）。multipart：audio 文件 + video_id/sentence_index/text。
func (h *Handler) Evaluate(c *gin.Context) {
	var in EvaluateInput
	if err := c.ShouldBind(&in); err != nil {
		response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))
		return
	}
	fileHeader, err := c.FormFile("audio")
	if err != nil {
		response.Error(c, errcode.ErrParam.WithMsg("缺少音频文件 audio"))
		return
	}
	if fileHeader.Size > maxAudioBytes {
		response.Error(c, errcode.ErrAudioFormat.WithMsg("音频文件过大(上限8MB)"))
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		response.Error(c, errcode.ErrServer)
		return
	}
	defer func() { _ = f.Close() }()
	audio, err := io.ReadAll(f)
	if err != nil {
		response.Error(c, errcode.ErrServer)
		return
	}

	uid := c.GetUint(middleware.ContextUserID)
	res, err := h.svc.Evaluate(c.Request.Context(), uid, in, audio)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, res)
}

// Records 训练历史（鉴权，分页）。
func (h *Handler) Records(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	uid := c.GetUint(middleware.ContextUserID)
	items, total, page, pageSize, err := h.svc.History(uid, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, items, total, page, pageSize)
}
