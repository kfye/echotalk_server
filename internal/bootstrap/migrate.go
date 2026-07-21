package bootstrap

import (
	"gorm.io/gorm"

	"github.com/echotalk/echotalk_server/internal/module/content"
	"github.com/echotalk/echotalk_server/internal/module/course"
	"github.com/echotalk/echotalk_server/internal/module/payment"
	"github.com/echotalk/echotalk_server/internal/module/training"
	"github.com/echotalk/echotalk_server/internal/module/user"
)

// AutoMigrate 自动迁移全部数据表（开发期幂等执行）。
// 新增模型后在此登记。
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&user.User{},               // users 用户
		&content.Video{},           // videos 视频内容
		&training.TrainingRecord{}, // training_records 训练记录
		&payment.Product{},         // products 商品/SKU
		&payment.Order{},           // orders 订单
		&payment.Membership{},      // memberships 会员
		&course.CoursePlan{},       // course_plans 训练营
		&course.DailyLesson{},      // daily_lessons 每日课
		&course.LessonWord{},       // lesson_words 词卡
		&course.LessonPhrase{},     // lesson_phrases 句型卡
		&course.Enrollment{},       // enrollments 报名
		&course.Checkin{},          // course_checkins 打卡
	)
}
