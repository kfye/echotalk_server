// Package speech 语音能力网关：业务模块访问 ASR/评测/TTS 的唯一入口。
// 通过 Provider 接口屏蔽厂商差异，业务模块禁止直接依赖具体厂商实现。
package speech

import "context"

// EvaluateRequest 发音评测请求。音频须为 16K/16bit/单声道 WAV。
type EvaluateRequest struct {
	Audio    []byte
	Text     string // 朗读文本
	Language string // 如 en_us
}

// EvaluateResult 发音评测结果。
type EvaluateResult struct {
	Overall   float64     `json:"overall"`   // 总分
	Accuracy  float64     `json:"accuracy"`  // 发音
	Fluency   float64     `json:"fluency"`   // 流利度
	Integrity float64     `json:"integrity"` // 完整度
	Words     []WordScore `json:"words"`     // 单词级评分
	Degraded  bool        `json:"degraded"`  // 是否为厂商不可用时的降级兜底结果
}

// WordScore 单词级评分。
type WordScore struct {
	Word  string  `json:"word"`
	Score float64 `json:"score"`
}

// Provider 语音厂商抽象（评测/识别/合成），便于多厂商切换。
type Provider interface {
	Evaluate(ctx context.Context, req EvaluateRequest) (EvaluateResult, error)
	// ASR / TTS 接口后续按需扩展。
}
