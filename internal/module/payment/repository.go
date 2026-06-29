package payment

import (
	"errors"

	"gorm.io/gorm"
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
