package user

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// ErrDuplicatePhone 手机号唯一键冲突（并发注册竞态兜底）。
var ErrDuplicatePhone = errors.New("duplicate phone")

// Repository 用户数据访问。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建仓储。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 新建用户。命中手机号唯一键冲突时返回 ErrDuplicatePhone。
func (r *Repository) Create(u *User) error {
	err := r.db.Create(u).Error
	var myErr *mysql.MySQLError
	if errors.As(err, &myErr) && myErr.Number == 1062 {
		return ErrDuplicatePhone
	}
	return err
}

// FindByPhone 按手机号查询；不存在返回 (nil, nil)。
func (r *Repository) FindByPhone(phone string) (*User, error) {
	var u User
	err := r.db.Where("phone = ?", phone).First(&u).Error
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
