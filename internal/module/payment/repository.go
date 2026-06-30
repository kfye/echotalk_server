package payment

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository 支付/订单数据访问。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建仓储。
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// Create 新建订单。
func (r *Repository) Create(o *Order) error { return r.db.Create(o).Error }

// GetProductByID 按 ID 取商品；不存在返回 (nil, nil)，由业务层转 errcode。
func (r *Repository) GetProductByID(id uint) (*Product, error) {
	var p Product
	if err := r.db.First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// ListOnlineProducts 取全部上架商品，按 sort 倒序、id 升序。
func (r *Repository) ListOnlineProducts() ([]Product, error) {
	var list []Product
	err := r.db.Where("status = ?", ProductStatusOnline).
		Order("sort DESC, id ASC").
		Find(&list).Error
	return list, err
}

// CreateProduct 新建商品（管理端）。
func (r *Repository) CreateProduct(p *Product) error { return r.db.Create(p).Error }

// ListProducts 管理端分页列商品（含下架，软删自动排除）；status 非 nil 时按状态过滤。
func (r *Repository) ListProducts(status *int8, page, pageSize int) ([]Product, int64, error) {
	q := r.db.Model(&Product{})
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []Product
	err := q.Order("sort DESC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

// UpdateProduct 全字段保存商品（管理端编辑/改状态）。
func (r *Repository) UpdateProduct(p *Product) error { return r.db.Save(p).Error }

// DeleteProduct 软删商品（Product 带 DeletedAt，物理保留以供订单回显）。
func (r *Repository) DeleteProduct(id uint) error { return r.db.Delete(&Product{}, id).Error }

// GetOrderByNo 按业务订单号查订单；不存在返回 (nil, nil)。
func (r *Repository) GetOrderByNo(orderNo string) (*Order, error) {
	var o Order
	if err := r.db.Where("order_no = ?", orderNo).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

// UpdateOrder 全字段保存订单（确认支付时更新状态/交易号/支付时间）。
func (r *Repository) UpdateOrder(o *Order) error { return r.db.Save(o).Error }

// ListOrders 管理端分页列订单，联表带用户邮箱与商品名，按创建时间倒序。
// status / userID 非 nil 时分别过滤。products 用原始 JOIN（软删商品仍回显商品名）。
func (r *Repository) ListOrders(status *int8, userID *uint, page, pageSize int) ([]AdminOrderItem, int64, error) {
	q := r.db.Table("orders AS o").
		Joins("LEFT JOIN users u ON u.id = o.user_id").
		Joins("LEFT JOIN products p ON p.id = o.product_id")
	if status != nil {
		q = q.Where("o.status = ?", *status)
	}
	if userID != nil {
		q = q.Where("o.user_id = ?", *userID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []AdminOrderItem
	err := q.
		Select("o.id, o.order_no, o.user_id, u.email AS email, " +
			"o.product_id, p.name AS product_name, o.amount, o.duration_days, " +
			"o.channel, o.status, o.paid_at, o.created_at").
		Order("o.created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Scan(&items).Error
	return items, total, err
}

// ListMemberships 管理端分页列会员，联表带用户邮箱，按到期时间倒序。
// status 实时按 expire_at 算（覆盖沉睡过期用户）：1=有效, 0=已过期/未生效；status 非 nil 时按实时值过滤。
func (r *Repository) ListMemberships(status *int8, page, pageSize int) ([]AdminMembershipItem, int64, error) {
	now := time.Now()
	q := r.db.Table("memberships AS m").
		Joins("LEFT JOIN users u ON u.id = m.user_id")
	if status != nil {
		if *status == MembershipStatusActive {
			q = q.Where("m.status = ? AND m.expire_at > ?", MembershipStatusActive, now)
		} else {
			q = q.Where("NOT (m.status = ? AND m.expire_at > ?)", MembershipStatusActive, now)
		}
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []AdminMembershipItem
	err := q.
		Select("m.id, m.user_id, u.email AS email, "+
			"(m.status = ? AND m.expire_at > ?) AS status, "+
			"m.start_at, m.expire_at, m.source, m.last_order_id, m.created_at",
			MembershipStatusActive, now).
		Order("m.expire_at DESC, m.id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Scan(&items).Error
	return items, total, err
}

// ExpireIfNeeded 懒更新：库存为有效但已过期时，把 status 翻成未生效（带条件单行幂等更新）。
// 由读路径(门禁/会员查询)触发，best-effort；不满足条件则 no-op。
func (r *Repository) ExpireIfNeeded(m *Membership) error {
	if m == nil || m.Status != MembershipStatusActive || m.ExpireAt.After(time.Now()) {
		return nil
	}
	if err := r.db.Model(&Membership{}).
		Where("id = ? AND status = ?", m.ID, MembershipStatusActive).
		Update("status", MembershipStatusInactive).Error; err != nil {
		return err
	}
	m.Status = MembershipStatusInactive // 同步内存对象，调用方据此返回
	return nil
}

// GetOrderByNoForUpdate 在事务内按订单号取订单并加行锁(SELECT ... FOR UPDATE)，
// 保证「判 pending + 置 paid」原子，避免并发重复确认。不存在返回 (nil, nil)。
func (r *Repository) GetOrderByNoForUpdate(orderNo string) (*Order, error) {
	var o Order
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("order_no = ?", orderNo).First(&o).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

// GetMembershipByUserID 取用户会员（一用户一条）；不存在返回 (nil, nil)。
func (r *Repository) GetMembershipByUserID(userID uint) (*Membership, error) {
	var m Membership
	if err := r.db.Where("user_id = ?", userID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// SaveMembership 新建或更新会员（开通/续期后落库）。
func (r *Repository) SaveMembership(m *Membership) error { return r.db.Save(m).Error }

// Tx 在事务内执行 fn，fn 收到绑定该事务的子仓库；返回 error 自动回滚。
func (r *Repository) Tx(fn func(txRepo *Repository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(&Repository{db: tx})
	})
}

// UserExists 校验目标用户是否存在（手动发卡用，按表名轻查，不 import user 包）。
func (r *Repository) UserExists(userID uint) (bool, error) {
	var count int64
	if err := r.db.Table("users").Where("id = ?", userID).Limit(1).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
