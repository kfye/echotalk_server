package course

import "time"

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

// LessonVideo 课内教学视频（轻查 videos 表；gorm column 显式映射，避免命名策略歧义）。
type LessonVideo struct {
	ID            uint   `json:"id" gorm:"column:id"`
	Title         string `json:"title" gorm:"column:title"`
	CoverURL      string `json:"cover_url" gorm:"column:cover_url"`
	HLSURL        string `json:"hls_url" gorm:"column:hls_url"`
	SubtitleEnURL string `json:"subtitle_en_url" gorm:"column:subtitle_en_url"`
	SubtitleCnURL string `json:"subtitle_cn_url" gorm:"column:subtitle_cn_url"`
	Duration      int    `json:"duration" gorm:"column:duration"`
}

// LessonWordItem 词卡（详情展示；含复现标记）。
type LessonWordItem struct {
	Word            string `json:"word"`
	Meaning         string `json:"meaning"`
	Phonetic        string `json:"phonetic"`
	ContextSentence string `json:"context_sentence"`
	IsRecur         bool   `json:"is_recur"`  // 是否复现词
	RecurSeq        int    `json:"recur_seq"` // 第几次见
}

// LessonPhraseItem 句型卡（详情展示）。
type LessonPhraseItem struct {
	Phrase  string `json:"phrase"`
	Meaning string `json:"meaning"`
	Example string `json:"example"`
}

// LessonTask 每日任务项（核心步骤，供客户端渲染）。
type LessonTask struct {
	Code  string `json:"code"`  // listen/shadow/cards/speak
	Title string `json:"title"` // 任务标题
}

// LessonDetailResponse 今日课详情（报名+解锁后返回）。
type LessonDetailResponse struct {
	DayIndex    int                `json:"day_index"`   // 第几天
	Title       string             `json:"title"`       // 课标题
	Description string             `json:"description"` // 课简介
	LessonType  int8               `json:"lesson_type"` // 课类型 1工作日2周六3周日
	Video       *LessonVideo       `json:"video"`       // 教学视频(工作日课有，否则 null)
	Words       []LessonWordItem   `json:"words"`       // 词卡
	Phrases     []LessonPhraseItem `json:"phrases"`     // 句型卡
	Tasks       []LessonTask       `json:"tasks"`       // 核心任务清单
	CheckedIn   bool               `json:"checked_in"`  // 该天是否已打卡
}

// CheckinRequest 打卡请求（客户端完成核心步骤后调用）。
type CheckinRequest struct {
	Day int `json:"day" binding:"required,min=1"` // 打卡的第几天
}

// CheckinResponse 打卡结果。
type CheckinResponse struct {
	DayIndex     int  `json:"day_index"`     // 打卡的第几天
	CheckedIn    bool `json:"checked_in"`    // 是否已打卡(恒 true)
	CheckinCount int  `json:"checkin_count"` // 累计打卡天数
}

// toWordItem / toPhraseItem 模型转详情项。
func toWordItem(w *LessonWord) LessonWordItem {
	return LessonWordItem{
		Word:            w.Word,
		Meaning:         w.Meaning,
		Phonetic:        w.Phonetic,
		ContextSentence: w.ContextSentence,
		IsRecur:         w.IsRecur,
		RecurSeq:        w.RecurSeq,
	}
}

func toPhraseItem(p *LessonPhrase) LessonPhraseItem {
	return LessonPhraseItem{Phrase: p.Phrase, Meaning: p.Meaning, Example: p.Example}
}

// lessonTasks 按课类型生成核心任务清单；工作日课为跟读四步，周六/周日留待迭代2/3。
func lessonTasks(lessonType int8) []LessonTask {
	if lessonType == LessonTypeWeekday {
		return []LessonTask{
			{Code: "listen", Title: "精听（带字幕1遍→无字幕1遍）"},
			{Code: "shadow", Title: "逐句跟读 + 评分"},
			{Code: "cards", Title: "今日词句卡"},
			{Code: "speak", Title: "开口任务（用今日句型录30-60秒）"},
		}
	}
	return []LessonTask{}
}

// ---------------- 管理端 ----------------

// CourseInput 管理端训练营新增/编辑入参。
type CourseInput struct {
	Name        string `json:"name" binding:"required,max=128"`       // 训练营名称(必填)
	Description string `json:"description" binding:"max=512"`         // 简介
	CoverURL    string `json:"cover_url" binding:"max=255"`           // 封面
	TotalDays   int    `json:"total_days" binding:"omitempty,min=1"`  // 总天数(默认90)
	TotalWeeks  int    `json:"total_weeks" binding:"omitempty,min=1"` // 总周数(默认12)
	ProductID   uint   `json:"product_id"`                            // 关联SKU(迭代2)
	Status      *int8  `json:"status" binding:"omitempty,oneof=0 1"`  // 上下架 0草稿1上架(不传默认草稿)
	Sort        int    `json:"sort"`                                  // 排序
}

// CourseStatusInput 管理端上下架入参。
type CourseStatusInput struct {
	Status *int8 `json:"status" binding:"required,oneof=0 1"` // 目标状态 0草稿1上架
}

// AdminCourseItem 管理端训练营项（全字段）。
type AdminCourseItem struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CoverURL    string    `json:"cover_url"`
	TotalDays   int       `json:"total_days"`
	TotalWeeks  int       `json:"total_weeks"`
	ProductID   uint      `json:"product_id"`
	Status      int8      `json:"status"` // 0草稿1上架
	Sort        int       `json:"sort"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// toAdminCourseItem 训练营模型转管理端项。
func toAdminCourseItem(c *CoursePlan) AdminCourseItem {
	return AdminCourseItem{
		ID: c.ID, Name: c.Name, Description: c.Description, CoverURL: c.CoverURL,
		TotalDays: c.TotalDays, TotalWeeks: c.TotalWeeks, ProductID: c.ProductID,
		Status: c.Status, Sort: c.Sort, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

// WordInput 词卡入参（随 lesson 提交）。
type WordInput struct {
	Word            string `json:"word" binding:"required,max=64"`
	Meaning         string `json:"meaning" binding:"max=255"`
	Phonetic        string `json:"phonetic" binding:"max=64"`
	ContextSentence string `json:"context_sentence" binding:"max=512"`
	IsRecur         bool   `json:"is_recur"`
	RecurSeq        int    `json:"recur_seq"`
	Sort            int    `json:"sort"`
}

// PhraseInput 句型卡入参（随 lesson 提交）。
type PhraseInput struct {
	Phrase  string `json:"phrase" binding:"required,max=255"`
	Meaning string `json:"meaning" binding:"max=255"`
	Example string `json:"example" binding:"max=512"`
	Sort    int    `json:"sort"`
}

// LessonInput 管理端每日课新增/编辑入参（含词句卡数组）。
type LessonInput struct {
	DayIndex    int           `json:"day_index" binding:"required,min=1"`          // 第几天(必填)
	Week        int           `json:"week"`                                        // 第几周
	LessonType  int8          `json:"lesson_type" binding:"omitempty,oneof=1 2 3"` // 1工作日2周六3周日(0默认1)
	VideoID     uint          `json:"video_id"`                                    // 教学视频ID(工作日课)
	Title       string        `json:"title" binding:"required,max=128"`            // 课标题(必填)
	Description string        `json:"description" binding:"max=512"`               // 课简介
	Sort        int           `json:"sort"`                                        // 排序
	Status      *int8         `json:"status" binding:"omitempty,oneof=0 1"`        // 上下架(不传默认草稿)
	Words       []WordInput   `json:"words"`                                       // 词卡数组
	Phrases     []PhraseInput `json:"phrases"`                                     // 句型卡数组
}

// AdminLessonItem 管理端每日课项（含词句卡）。
type AdminLessonItem struct {
	ID          uint               `json:"id"`
	CourseID    uint               `json:"course_id"`
	DayIndex    int                `json:"day_index"`
	Week        int                `json:"week"`
	LessonType  int8               `json:"lesson_type"`
	VideoID     uint               `json:"video_id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Sort        int                `json:"sort"`
	Status      int8               `json:"status"`
	Words       []LessonWordItem   `json:"words"`
	Phrases     []LessonPhraseItem `json:"phrases"`
}

// EnrollRequest 管理端手动报名入参。
type EnrollRequest struct {
	UserID    uint   `json:"user_id" binding:"required"`   // 目标用户ID
	CourseID  uint   `json:"course_id" binding:"required"` // 训练营ID
	StartDate string `json:"start_date"`                   // 开营日 YYYY-MM-DD(不传默认今天)
}

// EnrollResponse 手动报名结果。
type EnrollResponse struct {
	EnrollmentID uint   `json:"enrollment_id"`
	UserID       uint   `json:"user_id"`
	CourseID     uint   `json:"course_id"`
	StartDate    string `json:"start_date"`
	Status       int8   `json:"status"`
}
