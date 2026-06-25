package payment

import "time"

// 会员状态。
const (
	MembershipStatusInactive int8 = 0 // 未生效/已过期
	MembershipStatusActive   int8 = 1 // 有效
)

// 会员开通来源。
const (
	MembershipSourceOrder  int8 = 1 // 订单开通
	MembershipSourceManual int8 = 2 // 手动发卡
)

// Membership 会员实体（一用户一条，按 ExpireAt 做服务端付费门禁）。
type Membership struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"uniqueIndex" json:"user_id"`    // 用户ID(唯一)
	Status      int8      `gorm:"default:0;index" json:"status"` // 会员状态
	StartAt     time.Time `json:"start_at"`                      // 生效时间
	ExpireAt    time.Time `gorm:"index" json:"expire_at"`        // 到期时间
	Source      int8      `gorm:"default:1" json:"source"`       // 开通来源
	LastOrderID uint      `json:"last_order_id"`                 // 最近订单ID
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (Membership) TableName() string { return "memberships" }
