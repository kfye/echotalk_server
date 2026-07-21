package course

// PlanItem 训练营列表项（报名前展示）。
type PlanItem struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`        // 训练营名称
	Description string `json:"description"` // 简介
	CoverURL    string `json:"cover_url"`   // 封面
	TotalDays   int    `json:"total_days"`  // 总天数
	TotalWeeks  int    `json:"total_weeks"` // 总周数
}

// toPlanItem 把训练营模型转列表项。
func toPlanItem(c *CoursePlan) PlanItem {
	return PlanItem{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		CoverURL:    c.CoverURL,
		TotalDays:   c.TotalDays,
		TotalWeeks:  c.TotalWeeks,
	}
}

// LessonBrief 今日课概要（我的训练营里展示）。
type LessonBrief struct {
	DayIndex   int    `json:"day_index"`   // 第几天
	Title      string `json:"title"`       // 课标题
	LessonType int8   `json:"lesson_type"` // 课类型 1工作日2周六3周日
	VideoID    uint   `json:"video_id"`    // 教学视频ID(工作日课)
	CheckedIn  bool   `json:"checked_in"`  // 今日是否已打卡
}

// MyCourseResponse 我的训练营（进度+今日课）。未报名时 enrolled=false，其余为零值。
type MyCourseResponse struct {
	Enrolled     bool         `json:"enrolled"`      // 是否已报名
	Course       *PlanItem    `json:"course"`        // 训练营信息(未报名为 null)
	StartDate    string       `json:"start_date"`    // 开营日 YYYY-MM-DD
	CurrentDay   int          `json:"current_day"`   // 当前解锁到第几天
	CheckinCount int          `json:"checkin_count"` // 累计打卡天数
	Status       int8         `json:"status"`        // 报名状态 1进行2完成3过期
	TodayLesson  *LessonBrief `json:"today_lesson"`  // 今日课概要(无则 null)
}
