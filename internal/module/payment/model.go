package payment

import "time"

// 订单状态。
const (
	OrderStatusPending  int8 = 0 // 待支付
	OrderStatusPaid     int8 = 1 // 已支付
	OrderStatusRefunded int8 = 2 // 已退款
	OrderStatusClosed   int8 = 3 // 已关闭
)

// Order 订单实体（独立表边界）。
type Order struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	OrderNo   string     `gorm:"size:64;uniqueIndex" json:"order_no"` // 业务订单号
	UserID    uint       `gorm:"index" json:"user_id"`                // 用户ID
	ProductID uint       `gorm:"index" json:"product_id"`             // 商品ID
	Amount    int64      `json:"amount"`                              // 订单金额(分)
	Channel   string     `gorm:"size:32" json:"channel"`              // 支付渠道 mock/wechat/alipay
	Status    int8       `gorm:"default:0;index" json:"status"`       // 订单状态
	TradeNo   string     `gorm:"size:128" json:"trade_no"`            // 渠道交易号
	PaidAt    *time.Time `json:"paid_at"`                             // 支付时间
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// TableName 指定表名。
func (Order) TableName() string { return "orders" }
