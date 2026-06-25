package payment

import "gorm.io/gorm"

// Repository 支付/订单数据访问。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建仓储。
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// Create 新建订单。
func (r *Repository) Create(o *Order) error { return r.db.Create(o).Error }
