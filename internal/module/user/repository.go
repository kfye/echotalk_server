package user

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// ErrDuplicateEmail 邮箱唯一键冲突（并发注册竞态兜底）。
var ErrDuplicateEmail = errors.New("duplicate email")

// Repository 用户数据访问。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建仓储。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 新建用户。命中 email 唯一键冲突时返回 ErrDuplicateEmail。
func (r *Repository) Create(u *User) error {
	err := r.db.Create(u).Error
	var myErr *mysql.MySQLError
	if errors.As(err, &myErr) && myErr.Number == 1062 {
		return ErrDuplicateEmail
	}
	return err
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
