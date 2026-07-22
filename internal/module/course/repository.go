package course

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// ErrDuplicateCheckin 打卡唯一键冲突（并发重复打卡兜底）。
var ErrDuplicateCheckin = errors.New("duplicate checkin")

// ErrDuplicateEnrollment 报名唯一键冲突（一人一营）。
var ErrDuplicateEnrollment = errors.New("duplicate enrollment")

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

// ListWordsByLesson 取某课词卡，按 sort 升序、id 升序。
func (r *Repository) ListWordsByLesson(lessonID uint) ([]LessonWord, error) {
	var list []LessonWord
	err := r.db.Where("lesson_id = ?", lessonID).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

// ListPhrasesByLesson 取某课句型卡，按 sort 升序、id 升序。
func (r *Repository) ListPhrasesByLesson(lessonID uint) ([]LessonPhrase, error) {
	var list []LessonPhrase
	err := r.db.Where("lesson_id = ?", lessonID).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

// GetVideoForLesson 轻查 videos 表取课内教学视频字段；videoID=0 或不存在/已删返回 (nil, nil)。
// 直查表名（仿 payment.UserExists），不 import content，避免跨模块耦合。
func (r *Repository) GetVideoForLesson(videoID uint) (*LessonVideo, error) {
	if videoID == 0 {
		return nil, nil
	}
	var v LessonVideo
	err := r.db.Table("videos").
		Select("id, title, cover_url, hls_url, subtitle_en_url, subtitle_cn_url, duration").
		Where("id = ? AND deleted_at IS NULL", videoID).
		Limit(1).Scan(&v).Error
	if err != nil {
		return nil, err
	}
	if v.ID == 0 { // 未命中
		return nil, nil
	}
	return &v, nil
}

// CreateCheckin 写打卡记录；命中唯一键(报名+天)冲突返回 ErrDuplicateCheckin。
func (r *Repository) CreateCheckin(c *Checkin) error {
	err := r.db.Create(c).Error
	var myErr *mysql.MySQLError
	if errors.As(err, &myErr) && myErr.Number == 1062 {
		return ErrDuplicateCheckin
	}
	return err
}

// IncrCheckinCount 累计打卡数 +1。
func (r *Repository) IncrCheckinCount(enrollmentID uint) error {
	return r.db.Model(&Enrollment{}).Where("id = ?", enrollmentID).
		UpdateColumn("checkin_count", gorm.Expr("checkin_count + 1")).Error
}

// Tx 在事务内执行 fn，fn 收到绑定该事务的子仓库；返回 error 自动回滚。
func (r *Repository) Tx(fn func(txRepo *Repository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(&Repository{db: tx})
	})
}

// ---------------- 管理端 ----------------

// CreateCourse 新建训练营。
func (r *Repository) CreateCourse(c *CoursePlan) error { return r.db.Create(c).Error }

// ListCoursesAdmin 管理端分页列训练营（含草稿，软删自动排除）；status 非 nil 时过滤。
func (r *Repository) ListCoursesAdmin(status *int8, page, pageSize int) ([]CoursePlan, int64, error) {
	q := r.db.Model(&CoursePlan{})
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []CoursePlan
	err := q.Order("sort DESC, id ASC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

// UpdateCourse 全字段保存训练营。
func (r *Repository) UpdateCourse(c *CoursePlan) error { return r.db.Save(c).Error }

// DeleteCourse 软删训练营。
func (r *Repository) DeleteCourse(id uint) error { return r.db.Delete(&CoursePlan{}, id).Error }

// GetLessonByID 按 ID 取每日课；不存在返回 (nil, nil)。
func (r *Repository) GetLessonByID(id uint) (*DailyLesson, error) {
	var l DailyLesson
	if err := r.db.First(&l, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

// ListLessonsByCourse 取某训练营全部课，按 day_index 升序。
func (r *Repository) ListLessonsByCourse(courseID uint) ([]DailyLesson, error) {
	var list []DailyLesson
	err := r.db.Where("course_id = ?", courseID).Order("day_index ASC").Find(&list).Error
	return list, err
}

// CreateLessonWithCards 事务内建课 + 词句卡。
func (r *Repository) CreateLessonWithCards(l *DailyLesson, words []LessonWord, phrases []LessonPhrase) error {
	return r.Tx(func(tx *Repository) error {
		if err := tx.db.Create(l).Error; err != nil {
			return err
		}
		return tx.saveCards(l.ID, words, phrases)
	})
}

// UpdateLessonWithCards 事务内保存课 + 用新词句卡整替换旧的。
func (r *Repository) UpdateLessonWithCards(l *DailyLesson, words []LessonWord, phrases []LessonPhrase) error {
	return r.Tx(func(tx *Repository) error {
		if err := tx.db.Save(l).Error; err != nil {
			return err
		}
		if err := tx.db.Where("lesson_id = ?", l.ID).Delete(&LessonWord{}).Error; err != nil {
			return err
		}
		if err := tx.db.Where("lesson_id = ?", l.ID).Delete(&LessonPhrase{}).Error; err != nil {
			return err
		}
		return tx.saveCards(l.ID, words, phrases)
	})
}

// saveCards 批量写词句卡（lessonID 回填）。
func (r *Repository) saveCards(lessonID uint, words []LessonWord, phrases []LessonPhrase) error {
	for i := range words {
		words[i].LessonID = lessonID
		if err := r.db.Create(&words[i]).Error; err != nil {
			return err
		}
	}
	for i := range phrases {
		phrases[i].LessonID = lessonID
		if err := r.db.Create(&phrases[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

// DeleteLesson 软删每日课（词句卡保留，随课不再展示）。
func (r *Repository) DeleteLesson(id uint) error { return r.db.Delete(&DailyLesson{}, id).Error }

// UserExists 校验目标用户是否存在（手动报名用，轻查表名，不 import user 包）。
func (r *Repository) UserExists(userID uint) (bool, error) {
	var count int64
	if err := r.db.Table("users").Where("id = ? AND deleted_at IS NULL", userID).Limit(1).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateEnrollment 建报名；命中唯一键(用户+营)冲突返回 ErrDuplicateEnrollment。
func (r *Repository) CreateEnrollment(e *Enrollment) error {
	err := r.db.Create(e).Error
	var myErr *mysql.MySQLError
	if errors.As(err, &myErr) && myErr.Number == 1062 {
		return ErrDuplicateEnrollment
	}
	return err
}
