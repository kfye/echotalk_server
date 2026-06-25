package codesender

import "context"

// MockSender 内测验证码实现：发码为空操作，校验固定码。
type MockSender struct {
	fixedCode string
}

// NewMockSender 创建内测发码器，固定码 123456。
func NewMockSender() *MockSender {
	return &MockSender{fixedCode: "123456"}
}

// 编译期断言：确保实现了 CodeSender 接口。
var _ CodeSender = (*MockSender)(nil)

// Send 内测为空操作，直接成功。
func (m *MockSender) Send(ctx context.Context, target string) error {
	return nil
}

// Verify 比对固定码。
func (m *MockSender) Verify(ctx context.Context, target, code string) bool {
	return code == m.fixedCode
}
