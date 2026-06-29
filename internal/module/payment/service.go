package payment

import (
	"context"
	"fmt"
	"math/rand"
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

// ListProducts 取上架商品列表，供付费墙展示。
func (s *Service) ListProducts(ctx context.Context) ([]ProductItem, error) {
	products, err := s.repo.ListOnlineProducts()
	if err != nil {
		return nil, errcode.ErrServer
	}
	items := make([]ProductItem, 0, len(products))
	for i := range products {
		items = append(items, toProductItem(&products[i]))
	}
	return items, nil
}

// CreateOrder 校验商品后创建待支付订单，并向当前渠道取唤起支付所需参数。
// 仅落库 pending，不开会员（开会员在确认支付步）。
func (s *Service) CreateOrder(ctx context.Context, userID uint, req CreateOrderRequest) (*CreateOrderResponse, error) {
	product, err := s.repo.GetProductByID(req.ProductID)
	if err != nil {
		return nil, errcode.ErrServer
	}
	if product == nil {
		return nil, errcode.ErrProductNotFound
	}
	if product.Status != ProductStatusOnline {
		return nil, errcode.ErrProductOffline
	}

	// 价格与会员时长在下单时从商品快照，确认支付时据此开/续会员。
	order := &Order{
		OrderNo:      genOrderNo(),
		UserID:       userID,
		ProductID:    product.ID,
		Amount:       product.Price,
		DurationDays: product.DurationDays,
		Channel:      s.channel.Name(),
		Status:       OrderStatusPending,
	}
	if err := s.repo.Create(order); err != nil {
		return nil, errcode.ErrServer
	}

	// 取唤起支付参数（占位）；真实交易号在确认回调时落库，此处不置 paid。
	res, err := s.channel.Pay(ctx, channel.PayRequest{
		OrderNo: order.OrderNo,
		Amount:  order.Amount,
		Subject: product.Name,
		UserID:  userID,
	})
	if err != nil || !res.Success {
		return nil, errcode.ErrPayFailed
	}

	return &CreateOrderResponse{
		OrderNo:    order.OrderNo,
		ProductID:  order.ProductID,
		Amount:     order.Amount,
		Status:     order.Status,
		PayPayload: res.PayPayload,
	}, nil
}

// genOrderNo 生成业务订单号：ODR + 时间(到秒) + 4位随机，避免同秒并发撞唯一索引。
func genOrderNo() string {
	return fmt.Sprintf("ODR%s%04d", time.Now().Format("20060102150405"), rand.Intn(10000))
}
