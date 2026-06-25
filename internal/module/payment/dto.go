package payment

// CreateOrderRequest 创建订单请求。
type CreateOrderRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
}
