package payment

// CreateOrderRequest 创建订单请求。
type CreateOrderRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
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
