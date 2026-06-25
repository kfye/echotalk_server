package training

import "time"

// TrainingRecord 训练（跟读评测）记录。
// 因不再设 sentences 表，句子身份由字幕文件中的序号 SentenceIndex 标识，
// 并冗余存 SentenceText 作为评测试题文本，使记录自包含。
type TrainingRecord struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"index" json:"user_id"`                     // 用户ID
	VideoID        uint      `gorm:"index" json:"video_id"`                    // 视频ID
	SentenceIndex  int       `json:"sentence_index"`                           // 句在字幕文件中的序号
	SentenceText   string    `gorm:"size:512" json:"sentence_text"`            // 评测试题文本(冗余)
	AudioURL       string    `gorm:"size:255" json:"audio_url"`                // 用户录音地址(COS)
	OverallScore   float64   `gorm:"type:decimal(5,2)" json:"overall_score"`   // 总分
	AccuracyScore  float64   `gorm:"type:decimal(5,2)" json:"accuracy_score"`  // 发音准确度
	FluencyScore   float64   `gorm:"type:decimal(5,2)" json:"fluency_score"`   // 流利度
	IntegrityScore float64   `gorm:"type:decimal(5,2)" json:"integrity_score"` // 完整度
	WordDetails    string    `gorm:"type:json" json:"word_details"`            // 词级评分明细(JSON)
	Degraded       bool      `gorm:"default:false" json:"degraded"`            // 是否为讯飞降级兜底结果
	CreatedAt      time.Time `json:"created_at"`
}

// TableName 指定表名。
func (TrainingRecord) TableName() string { return "training_records" }
