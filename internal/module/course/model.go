package course

import (
	"time"

	"gorm.io/gorm"
)

// 训练营/每日课上下架状态（course 与 lesson 共用）。
const (
	StatusDraft  int8 = 0 // 草稿
	StatusOnline int8 = 1 // 上架
)

// 每日课类型。
const (
	LessonTypeWeekday  int8 = 1 // 工作日新课（周一~五，含教学视频）
	LessonTypeSaturday int8 = 2 // 周六场景强化（AI 对话，迭代3）
	LessonTypeSunday   int8 = 3 // 周日复习测试（听力快测，迭代2）
)

// 报名状态。
const (
	EnrollmentStatusActive    int8 = 1 // 进行中
	EnrollmentStatusCompleted int8 = 2 // 已完成
	EnrollmentStatusExpired   int8 = 3 // 已过期
)

// CoursePlan 训练营/课程计划（一个训练营=一份 90 天课表）。
type CoursePlan struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:128" json:"name"`          // 训练营名称
	Description string         `gorm:"size:512" json:"description"`   // 简介
	CoverURL    string         `gorm:"size:255" json:"cover_url"`     // 封面地址(CDN)
	TotalDays   int            `gorm:"default:90" json:"total_days"`  // 总天数(解锁上限)
	TotalWeeks  int            `gorm:"default:12" json:"total_weeks"` // 总周数
	ProductID   uint           `gorm:"index" json:"product_id"`       // 关联训练营SKU(迭代2用，可空)
	Status      int8           `gorm:"default:0;index" json:"status"` // 上下架状态 0草稿1上架
	Sort        int            `gorm:"default:0" json:"sort"`         // 排序权重
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"` // 软删除
}

// TableName 指定表名。
func (CoursePlan) TableName() string { return "course_plans" }

// DailyLesson 每日课（挂在某训练营下，按 DayIndex 解锁）。
type DailyLesson struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CourseID    uint           `gorm:"uniqueIndex:idx_course_day;index" json:"course_id"` // 所属训练营
	DayIndex    int            `gorm:"uniqueIndex:idx_course_day" json:"day_index"`       // 全局第几天(1..TotalDays，解锁基准)
	Week        int            `json:"week"`                                              // 第几周(1..12)
	LessonType  int8           `gorm:"default:1" json:"lesson_type"`                      // 课类型 1工作日2周六3周日
	VideoID     uint           `gorm:"index" json:"video_id"`                             // 关联教学视频(content.videos，工作日课有，可空)
	Title       string         `gorm:"size:128" json:"title"`                             // 课标题
	Description string         `gorm:"size:512" json:"description"`                       // 课简介
	Sort        int            `gorm:"default:0" json:"sort"`                             // 排序权重
	Status      int8           `gorm:"default:0;index" json:"status"`                     // 上下架状态 0草稿1上架
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"` // 软删除
}

// TableName 指定表名。
func (DailyLesson) TableName() string { return "daily_lessons" }

// LessonWord 词卡（当日新词，出自当日视频语境；含间隔重复复现标记）。
type LessonWord struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	LessonID        uint      `gorm:"index" json:"lesson_id"`           // 所属每日课
	Word            string    `gorm:"size:64" json:"word"`              // 单词
	Meaning         string    `gorm:"size:255" json:"meaning"`          // 释义
	Phonetic        string    `gorm:"size:64" json:"phonetic"`          // 音标(可选)
	ContextSentence string    `gorm:"size:512" json:"context_sentence"` // 语境例句(出自当日视频)
	IsRecur         bool      `gorm:"default:false" json:"is_recur"`    // 是否复现词(前几日出现过)
	RecurSeq        int       `gorm:"default:0" json:"recur_seq"`       // 第几次见到该词(复现次序)
	Sort            int       `gorm:"default:0" json:"sort"`            // 排序
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (LessonWord) TableName() string { return "lesson_words" }

// LessonPhrase 句型卡（当日新句型）。
type LessonPhrase struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	LessonID  uint      `gorm:"index" json:"lesson_id"`  // 所属每日课
	Phrase    string    `gorm:"size:255" json:"phrase"`  // 句型
	Meaning   string    `gorm:"size:255" json:"meaning"` // 释义/用法
	Example   string    `gorm:"size:512" json:"example"` // 例句
	Sort      int       `gorm:"default:0" json:"sort"`   // 排序
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (LessonPhrase) TableName() string { return "lesson_phrases" }

// Enrollment 报名（用户的训练营实例：起始日+进度+打卡）。一人一营一条。
type Enrollment struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"uniqueIndex:idx_user_course;index" json:"user_id"` // 用户ID
	CourseID     uint      `gorm:"uniqueIndex:idx_user_course" json:"course_id"`     // 训练营ID
	StartDate    time.Time `gorm:"type:date" json:"start_date"`                      // 开营日(解锁基准)
	Status       int8      `gorm:"default:1;index" json:"status"`                    // 报名状态 1进行2完成3过期
	CheckinCount int       `gorm:"default:0" json:"checkin_count"`                   // 累计打卡天数
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (Enrollment) TableName() string { return "enrollments" }

// Checkin 打卡记录（每报名每天一次）。
type Checkin struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	EnrollmentID uint      `gorm:"uniqueIndex:idx_enroll_day;index" json:"enrollment_id"` // 所属报名
	DayIndex     int       `gorm:"uniqueIndex:idx_enroll_day" json:"day_index"`           // 打卡的第几天
	CheckinDate  time.Time `gorm:"type:date" json:"checkin_date"`                         // 打卡日期
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名。
func (Checkin) TableName() string { return "course_checkins" }
