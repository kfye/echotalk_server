package content

import (
	"errors"

	"gorm.io/gorm"
)

// Repository 内容数据访问。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建仓储。
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// ListFilter 列表过滤条件。
type ListFilter struct {
	OnlineOnly bool  // true=仅上架(App)；false=全部(admin)
	Status     *int8 // admin 按状态过滤，可空
	Category   string
	Difficulty int8 // 0=不过滤
	Page       int
	PageSize   int
}

// List 分页查询，返回当页数据与总数。
func (r *Repository) List(f ListFilter) ([]Video, int64, error) {
	q := r.db.Model(&Video{})
	if f.OnlineOnly {
		q = q.Where("status = ?", VideoStatusOnline)
	} else if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	if f.Category != "" {
		q = q.Where("category = ?", f.Category)
	}
	if f.Difficulty > 0 {
		q = q.Where("difficulty = ?", f.Difficulty)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []Video
	err := q.Order("sort DESC, id DESC").
		Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).
		Find(&list).Error
	return list, total, err
}

// GetByID 按主键查询；不存在返回 (nil, nil)。
func (r *Repository) GetByID(id uint) (*Video, error) {
	var v Video
	err := r.db.First(&v, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// Create 新建视频。
func (r *Repository) Create(v *Video) error { return r.db.Create(v).Error }

// Update 全量更新视频字段。
func (r *Repository) Update(v *Video) error { return r.db.Save(v).Error }

// Delete 软删除。
func (r *Repository) Delete(id uint) error { return r.db.Delete(&Video{}, id).Error }
