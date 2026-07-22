package course

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/echotalk/echotalk_server/internal/middleware"
	"github.com/echotalk/echotalk_server/internal/pkg/jwt"
)

// RegisterRoutes 装配 course 模块路由。
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.Manager) {
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)

	// App 端：训练营（需鉴权）
	g := rg.Group("/course")
	g.Use(middleware.Auth(jwtManager))
	{
		g.GET("/plans", h.Plans)
		g.GET("/my", h.My)
		g.GET("/lessons/:day", h.LessonDetail)
		g.POST("/checkin", h.Checkin)
	}

	// 管理端：训练营 + 课 CRUD（先只校验登录，角色留后续）
	adminCourse := rg.Group("/admin/courses")
	adminCourse.Use(middleware.Auth(jwtManager))
	{
		adminCourse.GET("", h.AdminCourses)
		adminCourse.POST("", h.AdminCreateCourse)
		adminCourse.PUT("/:id", h.AdminUpdateCourse)
		adminCourse.PATCH("/:id/status", h.AdminUpdateCourseStatus)
		adminCourse.DELETE("/:id", h.AdminDeleteCourse)
		adminCourse.GET("/:id/lessons", h.AdminLessons)
		adminCourse.POST("/:id/lessons", h.AdminCreateLesson)
	}

	// 管理端：每日课编辑/删除
	adminLesson := rg.Group("/admin/lessons")
	adminLesson.Use(middleware.Auth(jwtManager))
	{
		adminLesson.PUT("/:id", h.AdminUpdateLesson)
		adminLesson.DELETE("/:id", h.AdminDeleteLesson)
	}

	// 管理端：手动报名（独立分组，避开 /admin/courses/:id 与静态段的 gin 路由冲突）
	adminEnroll := rg.Group("/admin/enrollments")
	adminEnroll.Use(middleware.Auth(jwtManager))
	{
		adminEnroll.POST("", h.AdminEnroll)
	}
}
