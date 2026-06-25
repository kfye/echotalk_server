// Package iflytek 科大讯飞 ISE 云端 API 的 Provider 实现。
package iflytek

import (
	"context"

	"github.com/echotalk/echotalk_server/internal/config"
	"github.com/echotalk/echotalk_server/internal/speech"
)

// ISEProvider 实现 speech.Provider，对接讯飞 ISE 云端评测 API。
type ISEProvider struct {
	cfg config.IflytekConfig
}

// NewISEProvider 创建讯飞 ISE Provider。
func NewISEProvider(cfg config.IflytekConfig) *ISEProvider {
	return &ISEProvider{cfg: cfg}
}

// 编译期断言：确保实现了 speech.Provider 接口。
var _ speech.Provider = (*ISEProvider)(nil)

// Evaluate 调用讯飞 ISE 评测。
func (p *ISEProvider) Evaluate(ctx context.Context, req speech.EvaluateRequest) (speech.EvaluateResult, error) {
	// TODO: 接入讯飞 ISE 云端 API（HMAC 鉴权签名 + WebSocket 流式上传音频，解析 XML 评测结果）。
	// 当前为骨架占位，返回空结果。
	return speech.EvaluateResult{}, nil
}
