package codesender

import "context"

// MockSender 内测送达渠道：返回固定码，不真正发送。
// 换真实短信/邮件时，只需新增一个实现把这两步换成「生成随机码 + 真正发送」。
type MockSender struct {
	fixedCode string
}

// NewMockSender 创建内测发码器，固定码 123456。
func NewMockSender() *MockSender {
	return &MockSender{fixedCode: "123456"}
}

// 编译期断言：确保实现了 CodeSender 接口。
var _ CodeSender = (*MockSender)(nil)

// Send 内测不真正发送，直接返回固定码供上层落库。
func (m *MockSender) Send(ctx context.Context, target string) (string, error) {
	return m.fixedCode, nil
}
