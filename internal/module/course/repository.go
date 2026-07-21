package course

import (
	"errors"

	"gorm.io/gorm"
)

// Repository 训练营数据访问。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建仓储。
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// ListOnlineCourses 取上架训练营，按 sort 倒序、id 升序。
func (r *Repository) ListOnlineCourses() ([]CoursePlan, error) {
	var list []CoursePlan
	err := r.db.Where("status = ?", StatusOnline).
		Order("sort DESC, id ASC").
		Find(&list).Error
	return list, err
}

// GetCourseByID 按 ID 取训练营；不存在返回 (nil, nil)。
func (r *Repository) GetCourseByID(id uint) (*CoursePlan, error) {
	var c CoursePlan
	if err := r.db.First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// GetActiveEnrollmentByUser 取用户当前进行中的报名（最新一条）；无则 (nil, nil)。
func (r *Repository) GetActiveEnrollmentByUser(userID uint) (*Enrollment, error) {
	var e Enrollment
	err := r.db.Where("user_id = ? AND status = ?", userID, EnrollmentStatusActive).
		Order("id DESC").First(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// GetLessonByCourseDay 取某训练营第 day 天的课；不存在返回 (nil, nil)。
func (r *Repository) GetLessonByCourseDay(courseID uint, day int) (*DailyLesson, error) {
	var l DailyLesson
	err := r.db.Where("course_id = ? AND day_index = ?", courseID, day).First(&l).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// CheckinExists 判断某报名第 day 天是否已打卡。
func (r *Repository) CheckinExists(enrollmentID uint, day int) (bool, error) {
	var count int64
	err := r.db.Model(&Checkin{}).
		Where("enrollment_id = ? AND day_index = ?", enrollmentID, day).
		Limit(1).Count(&count).Error
	return count > 0, err
}
