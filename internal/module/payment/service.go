package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/echotalk/echotalk_server/internal/module/payment/channel"
	"github.com/echotalk/echotalk_server/internal/pkg/errcode"
)

// Service 支付/会员业务逻辑。会员与渠道解耦，仅依赖 PaymentChannel 接口。
type Service struct {
	repo    *Repository
	channel channel.PaymentChannel
}

// NewService 创建服务。
func NewService(repo *Repository, ch channel.PaymentChannel) *Service {
	return &Service{repo: repo, channel: ch}
}

// CreateOrder 创建订单并通过当前渠道发起支付。
func (s *Service) CreateOrder(ctx context.Context, userID uint, req CreateOrderRequest) (*Order, error) {
	order := &Order{
		OrderNo:   genOrderNo(),
		UserID:    userID,
		ProductID: req.ProductID,
		Amount:    100, // TODO: 从 products 表读取价格
		Channel:   s.channel.Name(),
		Status:    OrderStatusPending,
	}
	if err := s.repo.Create(order); err != nil {
		return nil, errcode.ErrServer
	}
	res, err := s.channel.Pay(ctx, channel.PayRequest{
		OrderNo: order.OrderNo,
		Amount:  order.Amount,
		Subject: fmt.Sprintf("product-%d", req.ProductID),
		UserID:  userID,
	})
	if err != nil || !res.Success {
		return nil, errcode.ErrPayFailed
	}
	order.TradeNo = res.TradeNo
	return order, nil
}

func genOrderNo() string {
	return "ODR" + time.Now().Format("20060102150405")
}
