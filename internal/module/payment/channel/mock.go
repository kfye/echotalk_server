package channel

import "context"

// MockChannel 内测模拟支付渠道：直接返回成功。
type MockChannel struct{}

// NewMockChannel 创建模拟渠道。
func NewMockChannel() *MockChannel { return &MockChannel{} }

// 编译期断言：确保实现了 PaymentChannel 接口。
var _ PaymentChannel = (*MockChannel)(nil)

func (m *MockChannel) Name() string { return "mock" }

func (m *MockChannel) Pay(ctx context.Context, req PayRequest) (PayResult, error) {
	return PayResult{Success: true, TradeNo: "mock-" + req.OrderNo, PayPayload: "mock-payload"}, nil
}

func (m *MockChannel) Query(ctx context.Context, orderNo string) (PayResult, error) {
	return PayResult{Success: true, TradeNo: "mock-" + orderNo}, nil
}

func (m *MockChannel) Refund(ctx context.Context, orderNo string, amount int64) error {
	return nil
}

func (m *MockChannel) VerifyCallback(ctx context.Context, payload []byte) (CallbackResult, error) {
	return CallbackResult{Paid: true}, nil
}
