package training

import (
	"errors"

	"gorm.io/gorm"
)

// Repository 训练记录数据访问。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建仓储。
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// Create 写入一条训练记录。
func (r *Repository) Create(rec *TrainingRecord) error { return r.db.Create(rec).Error }

// ListByUser 按用户分页查询训练记录（新到旧），返回当页数据与总数。
func (r *Repository) ListByUser(userID uint, page, pageSize int) ([]TrainingRecord, int64, error) {
	q := r.db.Model(&TrainingRecord{}).Where("user_id = ?", userID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []TrainingRecord
	err := q.Order("id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

// GetByID 按主键查询；不存在返回 (nil, nil)。
func (r *Repository) GetByID(id uint) (*TrainingRecord, error) {
	var rec TrainingRecord
	err := r.db.First(&rec, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rec, nil
}
