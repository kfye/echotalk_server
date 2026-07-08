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

// GrantMembershipRequest 手动发卡请求（运营给指定用户开/续会员）。
type GrantMembershipRequest struct {
	UserID       uint `json:"user_id" binding:"required"`             // 目标用户ID
	DurationDays int  `json:"duration_days" binding:"required,min=1"` // 发卡时长(天)
}

// GrantResponse 手动发卡响应：发卡后会员态。
type GrantResponse struct {
	UserID     uint           `json:"user_id"`    // 目标用户ID
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

// AdminProductItem 管理端商品项（全字段，含 App 端隐藏的 status/sort/时间）。
type AdminProductItem struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`           // 商品名称
	Description   string    `json:"description"`    // 描述
	Price         int64     `json:"price"`          // 现价(分)
	OriginalPrice int64     `json:"original_price"` // 原价(分)
	DurationDays  int       `json:"duration_days"`  // 会员时长(天)
	Type          int8      `json:"type"`           // 商品类型 1订阅/2内容包
	Status        int8      `json:"status"`         // 上下架状态 0下架/1上架
	Sort          int       `json:"sort"`           // 排序权重
	CreatedAt     time.Time `json:"created_at"`     // 创建时间
	UpdatedAt     time.Time `json:"updated_at"`     // 更新时间
}

// toAdminProductItem 把商品模型转管理端项。
func toAdminProductItem(p *Product) AdminProductItem {
	return AdminProductItem{
		ID:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		Price:         p.Price,
		OriginalPrice: p.OriginalPrice,
		DurationDays:  p.DurationDays,
		Type:          p.Type,
		Status:        p.Status,
		Sort:          p.Sort,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// ProductInput 管理端商品新增/编辑入参。
type ProductInput struct {
	Name          string `json:"name" binding:"required,max=64"`       // 商品名称(必填)
	Description   string `json:"description" binding:"max=255"`        // 描述
	Price         int64  `json:"price" binding:"min=0"`                // 现价(分)
	OriginalPrice int64  `json:"original_price" binding:"min=0"`       // 原价(分)
	DurationDays  int    `json:"duration_days" binding:"min=0"`        // 会员时长(天，订阅型应>0)
	Type          int8   `json:"type" binding:"omitempty,oneof=1 2"`   // 商品类型 1订阅/2内容包(0时默认1)
	Status        *int8  `json:"status" binding:"omitempty,oneof=0 1"` // 上下架 0下架/1上架(不传默认下架)
	Sort          int    `json:"sort"`                                 // 排序权重
}

// ProductStatusInput 管理端上下架入参。
type ProductStatusInput struct {
	Status *int8 `json:"status" binding:"required,oneof=0 1"` // 目标状态 0下架/1上架
}

// applyProductInput 把入参写入商品模型；Type 为 0 时补默认订阅，Status 仅在传入时覆盖。
func applyProductInput(p *Product, in ProductInput) *Product {
	p.Name = in.Name
	p.Description = in.Description
	p.Price = in.Price
	p.OriginalPrice = in.OriginalPrice
	p.DurationDays = in.DurationDays
	if in.Type != 0 {
		p.Type = in.Type
	} else if p.Type == 0 {
		p.Type = ProductTypeSubscription
	}
	if in.Status != nil {
		p.Status = *in.Status
	}
	p.Sort = in.Sort
	return p
}

// AdminOrderItem 管理端订单列表项（联表带用户手机号与商品名，便于运营查看）。
type AdminOrderItem struct {
	ID           uint       `json:"id"`
	OrderNo      string     `json:"order_no"`      // 业务订单号
	UserID       uint       `json:"user_id"`       // 用户ID
	Phone        string     `json:"phone"`         // 用户手机号(联表)
	ProductID    uint       `json:"product_id"`    // 商品ID
	ProductName  string     `json:"product_name"`  // 商品名(联表，软删商品仍回显)
	Amount       int64      `json:"amount"`        // 金额(分)
	DurationDays int        `json:"duration_days"` // 会员时长(天)
	Channel      string     `json:"channel"`       // 支付渠道 mock/wechat/alipay
	Status       int8       `json:"status"`        // 订单状态 0待支付 1已支付 2已退款 3已关闭
	PaidAt       *time.Time `json:"paid_at"`       // 支付时间(未支付为空)
	CreatedAt    time.Time  `json:"created_at"`    // 创建时间
}

// AdminMembershipItem 管理端会员列表项（联表带手机号；status 为 SQL 实时算的有效性）。
type AdminMembershipItem struct {
	ID          uint      `json:"id"`            // 会员记录ID
	UserID      uint      `json:"user_id"`       // 用户ID
	Phone       string    `json:"phone"`         // 用户手机号(联表)
	Status      int8      `json:"status"`        // 实时有效性 0已过期/未生效 1有效(SQL按expire_at算)
	StartAt     time.Time `json:"start_at"`      // 生效时间
	ExpireAt    time.Time `json:"expire_at"`     // 到期时间
	Source      int8      `json:"source"`        // 开通来源 1订单 2手动发卡
	LastOrderID uint      `json:"last_order_id"` // 最近订单ID
	CreatedAt   time.Time `json:"created_at"`    // 创建时间
}
