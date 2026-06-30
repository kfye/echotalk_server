package payment

import (
	"time"

	"gorm.io/gorm"
)

// 商品类型。
const (
	ProductTypeSubscription int8 = 1 // 订阅
	ProductTypePackage      int8 = 2 // 内容包
)

// 商品上下架状态。
const (
	ProductStatusOffline int8 = 0 // 下架
	ProductStatusOnline  int8 = 1 // 上架
)

// Product 商品 / SKU 实体。
type Product struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:64" json:"name"`           // 商品名称
	Description   string    `gorm:"size:255" json:"description"`   // 描述
	Price         int64     `json:"price"`                         // 现价(分)
	OriginalPrice int64     `json:"original_price"`                // 原价(分)
	DurationDays  int       `json:"duration_days"`                 // 会员时长(天)
	Type          int8      `gorm:"default:1" json:"type"`         // 商品类型
	Status        int8      `gorm:"default:0;index" json:"status"` // 上下架状态
	Sort          int       `gorm:"default:0" json:"sort"`         // 排序权重
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 软删除（运营删 SKU 不物理消失，保订单回显引用）
}

// TableName 指定表名。
func (Product) TableName() string { return "products" }
