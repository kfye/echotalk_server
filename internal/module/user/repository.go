package user

import (
	"errors"

	"gorm.io/gorm"
)

// Repository 用户数据访问。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建仓储。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 新建用户。
func (r *Repository) Create(u *User) error {
	return r.db.Create(u).Error
}

// FindByEmail 按邮箱查询；不存在返回 (nil, nil)。
func (r *Repository) FindByEmail(email string) (*User, error) {
	var u User
	err := r.db.Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByID 按主键查询；不存在返回 (nil, nil)。
func (r *Repository) FindByID(id uint) (*User, error) {
	var u User
	err := r.db.First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
