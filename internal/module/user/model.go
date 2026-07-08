package user

import (
	"time"

	"gorm.io/gorm"
)

// User 用户实体（独立表边界）。
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Phone     string         `gorm:"size:32;uniqueIndex" json:"phone"`      // 手机号(注册/登录唯一 identifier)
	Email     string         `gorm:"size:128;index" json:"email,omitempty"` // 邮箱(保留备用，非注册入口)
	Password  string         `gorm:"size:128" json:"-"`                     // 密码哈希(bcrypt)
	Nickname  string         `gorm:"size:64" json:"nickname"`               // 昵称
	Avatar    string         `gorm:"size:255" json:"avatar,omitempty"`      // 头像地址
	Status    int8           `gorm:"default:1;index" json:"status"`         // 账号状态 1正常 0禁用
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 软删除(账号注销)
}

// TableName 指定表名。
func (User) TableName() string { return "users" }
