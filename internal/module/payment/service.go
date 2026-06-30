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

// GrantMembership 运营手动发卡：给指定用户开/续会员（Source=Manual，无订单）。
// 复用 applyMembership 的开通/续期逻辑，与订单确认口径一致。
func (s *Service) GrantMembership(ctx context.Context, req GrantMembershipRequest) (*GrantResponse, error) {
	exists, err := s.repo.UserExists(req.UserID)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if !exists {
		return nil, errcode.ErrUserNotFound
	}

	var member *Membership
	txErr := s.repo.Tx(func(tx *Repository) error {
		m, err := tx.GetMembershipByUserID(req.UserID)
		if err != nil {
			return err
		}
		m = applyMembership(m, req.UserID, req.DurationDays, MembershipSourceManual, 0, time.Now())
		if err := tx.SaveMembership(m); err != nil {
			return err
		}
		member = m
		return nil
	})
	if txErr != nil {
		return nil, errcode.ErrServer.Wrap(txErr)
	}

	return &GrantResponse{UserID: req.UserID, Membership: toMembershipInfo(member)}, nil
}

// MembershipStatus 查会员状态（个人中心）。有效性实时按到期时间判定。
func (s *Service) MembershipStatus(userID uint) (MembershipStatusResponse, error) {
	m, err := s.repo.GetMembershipByUserID(userID)
	if err != nil {
		return MembershipStatusResponse{}, errcode.ErrServer.Wrap(err)
	}
	resp := MembershipStatusResponse{Status: MembershipStatusInactive}
	if m != nil {
		resp.StartAt = m.StartAt
		resp.ExpireAt = m.ExpireAt
		if m.Status == MembershipStatusActive && m.ExpireAt.After(time.Now()) {
			resp.IsMember = true
			resp.Status = MembershipStatusActive
		} else {
			_ = s.repo.ExpireIfNeeded(m) // 懒更新：库存有效但已过期则翻库存列，best-effort
		}
	}
	return resp, nil
}

// ListProducts 取上架商品列表，供付费墙展示。
func (s *Service) ListProducts(ctx context.Context) ([]ProductItem, error) {
	products, err := s.repo.ListOnlineProducts()
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	items := make([]ProductItem, 0, len(products))
	for i := range products {
		items = append(items, toProductItem(&products[i]))
	}
	return items, nil
}

// 管理端分页默认值。
const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

// normalizePage 规范化分页参数（默认 1/20，上限 100）。
func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = defaultPage
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

// AdminListProducts 管理端分页列商品（含下架）；status 非 nil 时按状态过滤。
func (s *Service) AdminListProducts(status *int8, page, pageSize int) ([]AdminProductItem, int64, int, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	products, total, err := s.repo.ListProducts(status, page, pageSize)
	if err != nil {
		return nil, 0, 0, 0, errcode.ErrServer.Wrap(err)
	}
	items := make([]AdminProductItem, 0, len(products))
	for i := range products {
		items = append(items, toAdminProductItem(&products[i]))
	}
	return items, total, page, pageSize, nil
}

// AdminListOrders 管理端分页列订单；status / userID 非 nil 时分别过滤。
func (s *Service) AdminListOrders(status *int8, userID *uint, page, pageSize int) ([]AdminOrderItem, int64, int, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	items, total, err := s.repo.ListOrders(status, userID, page, pageSize)
	if err != nil {
		return nil, 0, 0, 0, errcode.ErrServer.Wrap(err)
	}
	if items == nil {
		items = []AdminOrderItem{}
	}
	return items, total, page, pageSize, nil
}

// AdminListMemberships 管理端分页列会员；status 实时按到期时间算，非 nil 时按实时值过滤。
func (s *Service) AdminListMemberships(status *int8, page, pageSize int) ([]AdminMembershipItem, int64, int, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	items, total, err := s.repo.ListMemberships(status, page, pageSize)
	if err != nil {
		return nil, 0, 0, 0, errcode.ErrServer.Wrap(err)
	}
	if items == nil {
		items = []AdminMembershipItem{}
	}
	return items, total, page, pageSize, nil
}

// CreateProduct 新建 SKU；不传 status 默认下架（运营确认后再上架）。
func (s *Service) CreateProduct(in ProductInput) (*AdminProductItem, error) {
	p := applyProductInput(&Product{Status: ProductStatusOffline}, in)
	if err := s.repo.CreateProduct(p); err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	item := toAdminProductItem(p)
	return &item, nil
}

// UpdateProduct 编辑 SKU。
func (s *Service) UpdateProduct(id uint, in ProductInput) (*AdminProductItem, error) {
	p, err := s.repo.GetProductByID(id)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if p == nil {
		return nil, errcode.ErrProductNotFound
	}
	p = applyProductInput(p, in)
	if err := s.repo.UpdateProduct(p); err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	item := toAdminProductItem(p)
	return &item, nil
}

// UpdateProductStatus 上下架 SKU。
func (s *Service) UpdateProductStatus(id uint, status int8) (*AdminProductItem, error) {
	p, err := s.repo.GetProductByID(id)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if p == nil {
		return nil, errcode.ErrProductNotFound
	}
	p.Status = status
	if err := s.repo.UpdateProduct(p); err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	item := toAdminProductItem(p)
	return &item, nil
}

// DeleteProduct 软删 SKU（物理保留以供订单回显）。
func (s *Service) DeleteProduct(id uint) error {
	p, err := s.repo.GetProductByID(id)
	if err != nil {
		return errcode.ErrServer.Wrap(err)
	}
	if p == nil {
		return errcode.ErrProductNotFound
	}
	if err := s.repo.DeleteProduct(id); err != nil {
		return errcode.ErrServer.Wrap(err)
	}
	return nil
}

// CreateOrder 校验商品后创建待支付订单，并向当前渠道取唤起支付所需参数。
// 仅落库 pending，不开会员（开会员在确认支付步）。
func (s *Service) CreateOrder(ctx context.Context, userID uint, req CreateOrderRequest) (*CreateOrderResponse, error) {
	product, err := s.repo.GetProductByID(req.ProductID)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
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
		return nil, errcode.ErrServer.Wrap(err)
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

// ConfirmOrder 模拟支付确认：确认成功后置订单已支付并开通/续期会员（事务内）。
// 已支付订单幂等返回当前会员态，不重复开会员。
func (s *Service) ConfirmOrder(ctx context.Context, userID uint, orderNo string) (*ConfirmResponse, error) {
	order, err := s.repo.GetOrderByNo(orderNo)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	// 不存在或非本人：统一按"订单不存在"，防订单号枚举。
	if order == nil || order.UserID != userID {
		return nil, errcode.ErrOrderNotFound
	}

	// 幂等：已支付直接返回当前会员态，不重复开会员。
	if order.Status == OrderStatusPaid {
		m, err := s.repo.GetMembershipByUserID(userID)
		if err != nil {
			return nil, errcode.ErrServer.Wrap(err)
		}
		return &ConfirmResponse{OrderNo: order.OrderNo, Status: order.Status, Membership: toMembershipInfo(m)}, nil
	}
	// 退款/关闭等非待支付态：无法支付。
	if order.Status != OrderStatusPending {
		return nil, errcode.ErrOrderStatus
	}

	// 渠道查单确认（事务外，不持锁做网络调用）。
	res, err := s.channel.Query(ctx, orderNo)
	if err != nil || !res.Success {
		return nil, errcode.ErrPayFailed
	}

	// 事务内：行锁重查 + 再判 pending（防并发重复确认）→ 置 paid → 开/续会员。
	var member *Membership
	txErr := s.repo.Tx(func(tx *Repository) error {
		o, err := tx.GetOrderByNoForUpdate(orderNo)
		if err != nil {
			return err
		}
		if o == nil {
			return errcode.ErrOrderNotFound
		}
		if o.Status == OrderStatusPaid { // 并发下已被另一请求确认
			member, err = tx.GetMembershipByUserID(userID)
			return err
		}
		if o.Status != OrderStatusPending {
			return errcode.ErrOrderStatus
		}

		now := time.Now()
		o.Status = OrderStatusPaid
		o.PaidAt = &now
		o.TradeNo = res.TradeNo
		if err := tx.UpdateOrder(o); err != nil {
			return err
		}

		m, err := tx.GetMembershipByUserID(userID)
		if err != nil {
			return err
		}
		m = applyMembership(m, userID, o.DurationDays, MembershipSourceOrder, o.ID, now)
		if err := tx.SaveMembership(m); err != nil {
			return err
		}
		member = m
		order = o
		return nil
	})
	if txErr != nil {
		if be, ok := txErr.(*errcode.Error); ok {
			return nil, be
		}
		return nil, errcode.ErrServer.Wrap(txErr)
	}

	return &ConfirmResponse{OrderNo: order.OrderNo, Status: order.Status, Membership: toMembershipInfo(member)}, nil
}

// applyMembership 计算开通/续期后的会员：仍有效则到期日顺延，过期/新建则从现在重新起算。
// m 可为 nil（首次开通）。供订单确认与手动发卡(Task E)共用。
func applyMembership(m *Membership, userID uint, durationDays int, source int8, lastOrderID uint, now time.Time) *Membership {
	if m == nil {
		m = &Membership{UserID: userID, StartAt: now}
	}
	base := now
	if m.ExpireAt.After(now) { // 当前仍有效 → 顺延，StartAt 不变
		base = m.ExpireAt
	} else { // 过期或新建 → 重新起算
		m.StartAt = now
	}
	m.ExpireAt = base.AddDate(0, 0, durationDays)
	m.Status = MembershipStatusActive
	m.Source = source
	m.LastOrderID = lastOrderID
	return m
}

// genOrderNo 生成业务订单号：ODR + 时间(到秒) + 4位随机，避免同秒并发撞唯一索引。
func genOrderNo() string {
	return fmt.Sprintf("ODR%s%04d", time.Now().Format("20060102150405"), rand.Intn(10000))
}
