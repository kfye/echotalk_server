package payment

import "time"

// CreateOrderRequest 创建订单请求。
type CreateOrderRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
}

// CreateOrderResponse 下单响应：待支付订单 + 唤起支付所需参数。
type CreateOrderResponse struct {
	OrderNo    string `json:"order_no"`    // 业务订单号(确认支付时回传)
	ProductID  uint   `json:"product_id"`  // 商品ID
	Amount     int64  `json:"amount"`      // 应付金额(分)
	Status     int8   `json:"status"`      // 订单状态 0待支付
	PayPayload string `json:"pay_payload"` // 客户端唤起支付所需参数(mock占位)
}

// MembershipInfo 会员态（供确认响应与会员查询接口复用）。
type MembershipInfo struct {
	Status   int8      `json:"status"`    // 会员状态 0未生效/已过期 1有效
	StartAt  time.Time `json:"start_at"`  // 生效时间
	ExpireAt time.Time `json:"expire_at"` // 到期时间
}

// toMembershipInfo 把会员模型转会员态；m 为 nil 时返回未生效零态。
func toMembershipInfo(m *Membership) MembershipInfo {
	if m == nil {
		return MembershipInfo{Status: MembershipStatusInactive}
	}
	return MembershipInfo{Status: m.Status, StartAt: m.StartAt, ExpireAt: m.ExpireAt}
}

// ConfirmResponse 支付确认响应：订单已支付 + 最新会员态。
type ConfirmResponse struct {
	OrderNo    string         `json:"order_no"`   // 业务订单号
	Status     int8           `json:"status"`     // 订单状态 1已支付
	Membership MembershipInfo `json:"membership"` // 开通/续期后的会员态
}

// MembershipStatusResponse 会员状态查询响应（个人中心）。
type MembershipStatusResponse struct {
	IsMember bool      `json:"is_member"` // 当前是否有效会员(实时按到期时间)
	Status   int8      `json:"status"`    // 会员状态 0非会员/已过期 1有效
	StartAt  time.Time `json:"start_at"`  // 生效时间(无会员为零值)
	ExpireAt time.Time `json:"expire_at"` // 到期时间(无会员为零值，过期则为过去时间)
}

// ProductItem 商品/SKU 列表项（付费墙展示用，不外泄内部字段）。
type ProductItem struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`           // 商品名称
	Description   string `json:"description"`    // 描述
	Price         int64  `json:"price"`          // 现价(分)
	OriginalPrice int64  `json:"original_price"` // 原价(分)
	DurationDays  int    `json:"duration_days"`  // 会员时长(天)
	Type          int8   `json:"type"`           // 商品类型 1订阅/2内容包
}

// toProductItem 把商品模型转列表项。
func toProductItem(p *Product) ProductItem {
	return ProductItem{
		ID:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		Price:         p.Price,
		OriginalPrice: p.OriginalPrice,
		DurationDays:  p.DurationDays,
		Type:          p.Type,
	}
}
