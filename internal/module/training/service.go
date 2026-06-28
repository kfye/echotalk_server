package training

import (
	"context"
	"encoding/json"

	"github.com/echotalk/echotalk_server/internal/pkg/errcode"
	"github.com/echotalk/echotalk_server/internal/speech"
)

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
}

// NewService 创建服务。
func NewService(repo *Repository, gw *speech.Gateway) *Service {
	return &Service{repo: repo, speech: gw}
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
	rec := &TrainingRecord{
		UserID:         userID,
		VideoID:        in.VideoID,
		SentenceIndex:  in.SentenceIndex,
		SentenceText:   in.Text,
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
