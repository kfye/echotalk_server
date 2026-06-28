package training

import (
	"time"

	"github.com/echotalk/echotalk_server/internal/speech"
)

// EvaluateInput 评测接口入参（multipart/form-data；音频文件以 audio 字段单独取）。
type EvaluateInput struct {
	VideoID       uint   `form:"video_id" binding:"required"`     // 所属视频ID
	SentenceIndex int    `form:"sentence_index"`                  // 句在字幕文件中的序号，与字幕 seq 对齐
	Text          string `form:"text" binding:"required,max=512"` // 评测试题文本=句子英文
}

// EvaluateResponse 评测结果响应。
type EvaluateResponse struct {
	RecordID  uint               `json:"record_id"` // 落库后的训练记录ID
	Overall   float64            `json:"overall"`   // 总分
	Accuracy  float64            `json:"accuracy"`  // 发音准确度
	Fluency   float64            `json:"fluency"`   // 流利度
	Integrity float64            `json:"integrity"` // 完整度
	Words     []speech.WordScore `json:"words"`     // 单词级评分
	Degraded  bool               `json:"degraded"`  // 讯飞不可用时的降级兜底结果
}

// RecordItem 训练历史列表项（不含词级明细，详情再取）。
type RecordItem struct {
	ID             uint      `json:"id"`              // 训练记录ID
	VideoID        uint      `json:"video_id"`        // 所属视频ID
	SentenceIndex  int       `json:"sentence_index"`  // 句在字幕文件中的序号
	SentenceText   string    `json:"sentence_text"`   // 评测试题文本(冗余)
	OverallScore   float64   `json:"overall_score"`   // 总分
	AccuracyScore  float64   `json:"accuracy_score"`  // 发音准确度
	FluencyScore   float64   `json:"fluency_score"`   // 流利度
	IntegrityScore float64   `json:"integrity_score"` // 完整度
	Degraded       bool      `json:"degraded"`        // 是否为降级兜底结果
	CreatedAt      time.Time `json:"created_at"`      // 评测时间
}

// toRecordItem 把模型转历史列表项。
func toRecordItem(r *TrainingRecord) RecordItem {
	return RecordItem{
		ID:             r.ID,
		VideoID:        r.VideoID,
		SentenceIndex:  r.SentenceIndex,
		SentenceText:   r.SentenceText,
		OverallScore:   r.OverallScore,
		AccuracyScore:  r.AccuracyScore,
		FluencyScore:   r.FluencyScore,
		IntegrityScore: r.IntegrityScore,
		Degraded:       r.Degraded,
		CreatedAt:      r.CreatedAt,
	}
}
