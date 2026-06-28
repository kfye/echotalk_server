package training

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/echotalk/echotalk_server/internal/pkg/errcode"
	"github.com/echotalk/echotalk_server/internal/speech"
)

// AudioUploader 录音存储抽象（可空）。cos.Uploader 满足此接口。
// 设为接口以便 training 模块不直接耦合 COS 实现，且未配置时可传 nil。
type AudioUploader interface {
	UploadBytes(ctx context.Context, data []byte, objectKey string) (string, error)
}

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100

	// degradedMessage 讯飞不可用时返回给客户端的提示语。
	degradedMessage = "评测服务暂不可用，请稍后重试"
)

// Service 训练业务逻辑：评测编排 + 历史。
type Service struct {
	repo   *Repository
	speech *speech.Gateway
	audio  AudioUploader // 可空：录音存储，nil 时不存
	logger *zap.Logger
}

// NewService 创建服务。audio 可为 nil（未配置 COS 时录音不存储）。
func NewService(repo *Repository, gw *speech.Gateway, audio AudioUploader, logger *zap.Logger) *Service {
	return &Service{repo: repo, speech: gw, audio: audio, logger: logger}
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = defaultPage
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

// Evaluate 评测编排：录音送语音网关 → 落库 → 返回结构化评分。
// 网关已对厂商异常降级（Degraded），降级结果同样落库；仅音频格式错误等才返回 error。
func (s *Service) Evaluate(ctx context.Context, userID uint, in EvaluateInput, audio []byte) (*EvaluateResponse, error) {
	res, err := s.speech.Evaluate(ctx, speech.EvaluateRequest{
		Audio:    audio,
		Text:     in.Text,
		Language: "en_us",
	})
	if err != nil {
		return nil, err // 多为 errcode.ErrAudioFormat（400），直接透传
	}

	wordsJSON, err := json.Marshal(res.Words)
	if err != nil {
		return nil, errcode.ErrServer
	}

	// 录音存 COS（优雅降级）：未配置或上传失败不影响评测，audio_url 留空。
	audioURL := s.uploadAudio(ctx, userID, audio)

	rec := &TrainingRecord{
		UserID:         userID,
		VideoID:        in.VideoID,
		SentenceIndex:  in.SentenceIndex,
		SentenceText:   in.Text,
		AudioURL:       audioURL,
		OverallScore:   res.Overall,
		AccuracyScore:  res.Accuracy,
		FluencyScore:   res.Fluency,
		IntegrityScore: res.Integrity,
		WordDetails:    string(wordsJSON),
		Degraded:       res.Degraded,
	}
	if err := s.repo.Create(rec); err != nil {
		return nil, errcode.ErrServer
	}

	resp := &EvaluateResponse{
		RecordID:  rec.ID,
		Overall:   res.Overall,
		Accuracy:  res.Accuracy,
		Fluency:   res.Fluency,
		Integrity: res.Integrity,
		Words:     res.Words,
		Degraded:  res.Degraded,
	}
	if res.Degraded {
		resp.Message = degradedMessage
	}
	return resp, nil
}

// uploadAudio 把录音传 COS，返回可访问 URL；未配置上传器或失败时返回空串（不阻断评测）。
func (s *Service) uploadAudio(ctx context.Context, userID uint, audio []byte) string {
	if s.audio == nil {
		return ""
	}
	key := fmt.Sprintf("training/%d/%d.wav", userID, time.Now().UnixNano())
	url, err := s.audio.UploadBytes(ctx, audio, key)
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("training: 录音存 COS 失败，audio_url 留空", zap.Error(err))
		}
		return ""
	}
	return url
}

// History 训练历史分页（按当前用户）。
func (s *Service) History(userID uint, page, pageSize int) ([]RecordItem, int64, int, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	recs, total, err := s.repo.ListByUser(userID, page, pageSize)
	if err != nil {
		return nil, 0, 0, 0, errcode.ErrServer
	}
	items := make([]RecordItem, 0, len(recs))
	for i := range recs {
		items = append(items, toRecordItem(&recs[i]))
	}
	return items, total, page, pageSize, nil
}
