// Package channel 定义支付渠道适配器。内测用 MockChannel；
// 日后接微信/支付宝/IAP 只新增实现，会员逻辑不动。
package channel

import "context"

// PayRequest 发起支付请求。
type PayRequest struct {
	OrderNo string
	Amount  int64 // 单位：分
	Subject string
	UserID  uint
}

// PayResult 支付结果。
type PayResult struct {
	Success    bool
	TradeNo    string // 渠道交易号
	PayPayload string // 客户端唤起支付所需参数（占位）
}

// CallbackResult 支付回调解析结果。
type CallbackResult struct {
	OrderNo string
	TradeNo string
	Paid    bool
}

// PaymentChannel 支付渠道适配器接口。
type PaymentChannel interface {
	Name() string
	Pay(ctx context.Context, req PayRequest) (PayResult, error)
	Query(ctx context.Context, orderNo string) (PayResult, error)
	Refund(ctx context.Context, orderNo string, amount int64) error
	VerifyCallback(ctx context.Context, payload []byte) (CallbackResult, error)
}
