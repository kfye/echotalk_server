package course

import (
	"context"
	"errors"
	"time"

	"github.com/echotalk/echotalk_server/internal/pkg/errcode"
)

// Service 训练营业务逻辑。
type Service struct {
	repo *Repository
}

// NewService 创建服务。
func NewService(repo *Repository) *Service { return &Service{repo: repo} }

// ListPlans 取上架训练营列表（报名墙展示）。
func (s *Service) ListPlans(ctx context.Context) ([]PlanItem, error) {
	courses, err := s.repo.ListOnlineCourses()
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	items := make([]PlanItem, 0, len(courses))
	for i := range courses {
		items = append(items, toPlanItem(&courses[i]))
	}
	return items, nil
}

// MyCourse 我的训练营：报名进度 + 今日课概要。未报名返回 enrolled=false。
func (s *Service) MyCourse(userID uint) (MyCourseResponse, error) {
	e, err := s.repo.GetActiveEnrollmentByUser(userID)
	if err != nil {
		return MyCourseResponse{}, errcode.ErrServer.Wrap(err)
	}
	if e == nil {
		return MyCourseResponse{Enrolled: false}, nil
	}

	course, err := s.repo.GetCourseByID(e.CourseID)
	if err != nil {
		return MyCourseResponse{}, errcode.ErrServer.Wrap(err)
	}
	totalDays := 90
	if course != nil {
		totalDays = course.TotalDays
	}
	cur := unlockedDay(e.StartDate, totalDays, time.Now())

	resp := MyCourseResponse{
		Enrolled:     true,
		StartDate:    e.StartDate.Format("2006-01-02"),
		CurrentDay:   cur,
		CheckinCount: e.CheckinCount,
		Status:       e.Status,
	}
	if course != nil {
		pi := toPlanItem(course)
		resp.Course = &pi
	}

	// 今日课概要（当前解锁到的那天）
	if cur >= 1 {
		lesson, err := s.repo.GetLessonByCourseDay(e.CourseID, cur)
		if err != nil {
			return MyCourseResponse{}, errcode.ErrServer.Wrap(err)
		}
		if lesson != nil {
			checkedIn, err := s.repo.CheckinExists(e.ID, cur)
			if err != nil {
				return MyCourseResponse{}, errcode.ErrServer.Wrap(err)
			}
			resp.TodayLesson = &LessonBrief{
				DayIndex:   lesson.DayIndex,
				Title:      lesson.Title,
				LessonType: lesson.LessonType,
				VideoID:    lesson.VideoID,
				CheckedIn:  checkedIn,
			}
		}
	}
	return resp, nil
}

// LessonDetail 某天课详情：校验报名+解锁后返回 video+词句卡+任务+打卡态。
func (s *Service) LessonDetail(userID uint, day int) (*LessonDetailResponse, error) {
	e, err := s.repo.GetActiveEnrollmentByUser(userID)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if e == nil {
		return nil, errcode.ErrNotEnrolled
	}
	course, err := s.repo.GetCourseByID(e.CourseID)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	totalDays := 90
	if course != nil {
		totalDays = course.TotalDays
	}
	if !lessonUnlocked(day, unlockedDay(e.StartDate, totalDays, time.Now())) {
		return nil, errcode.ErrLessonLocked
	}

	lesson, err := s.repo.GetLessonByCourseDay(e.CourseID, day)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if lesson == nil {
		return nil, errcode.ErrLessonNotFound
	}

	words, err := s.repo.ListWordsByLesson(lesson.ID)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	phrases, err := s.repo.ListPhrasesByLesson(lesson.ID)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	video, err := s.repo.GetVideoForLesson(lesson.VideoID)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	checkedIn, err := s.repo.CheckinExists(e.ID, day)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}

	resp := &LessonDetailResponse{
		DayIndex:    lesson.DayIndex,
		Title:       lesson.Title,
		Description: lesson.Description,
		LessonType:  lesson.LessonType,
		Video:       video,
		Words:       make([]LessonWordItem, 0, len(words)),
		Phrases:     make([]LessonPhraseItem, 0, len(phrases)),
		Tasks:       lessonTasks(lesson.LessonType),
		CheckedIn:   checkedIn,
	}
	for i := range words {
		resp.Words = append(resp.Words, toWordItem(&words[i]))
	}
	for i := range phrases {
		resp.Phrases = append(resp.Phrases, toPhraseItem(&phrases[i]))
	}
	return resp, nil
}

// Checkin 打卡：校验报名+解锁+当天有课+未打卡 → 事务内写打卡并累计。幂等，重复 ErrAlreadyCheckedIn。
func (s *Service) Checkin(userID uint, day int) (*CheckinResponse, error) {
	e, err := s.repo.GetActiveEnrollmentByUser(userID)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if e == nil {
		return nil, errcode.ErrNotEnrolled
	}
	course, err := s.repo.GetCourseByID(e.CourseID)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	totalDays := 90
	if course != nil {
		totalDays = course.TotalDays
	}
	if !lessonUnlocked(day, unlockedDay(e.StartDate, totalDays, time.Now())) {
		return nil, errcode.ErrLessonLocked
	}
	lesson, err := s.repo.GetLessonByCourseDay(e.CourseID, day)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if lesson == nil {
		return nil, errcode.ErrLessonNotFound
	}

	// 事务内：写打卡（唯一键兜底重复）+ 累计 +1。
	txErr := s.repo.Tx(func(tx *Repository) error {
		if err := tx.CreateCheckin(&Checkin{EnrollmentID: e.ID, DayIndex: day, CheckinDate: time.Now()}); err != nil {
			if errors.Is(err, ErrDuplicateCheckin) {
				return errcode.ErrAlreadyCheckedIn
			}
			return err
		}
		return tx.IncrCheckinCount(e.ID)
	})
	if txErr != nil {
		if be, ok := txErr.(*errcode.Error); ok {
			return nil, be
		}
		return nil, errcode.ErrServer.Wrap(txErr)
	}

	return &CheckinResponse{DayIndex: day, CheckedIn: true, CheckinCount: e.CheckinCount + 1}, nil
}

// ---------------- 管理端 ----------------

// 分页默认值。
const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

// normalizePage 规范化分页参数（默认 1/20，上限 100）。
func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = defaultPage
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

// applyCourseInput 把入参写入训练营模型；天/周为 0 时补默认，status 仅传入时覆盖。
func applyCourseInput(c *CoursePlan, in CourseInput) {
	c.Name = in.Name
	c.Description = in.Description
	c.CoverURL = in.CoverURL
	c.TotalDays = in.TotalDays
	if c.TotalDays == 0 {
		c.TotalDays = 90
	}
	c.TotalWeeks = in.TotalWeeks
	if c.TotalWeeks == 0 {
		c.TotalWeeks = 12
	}
	c.ProductID = in.ProductID
	c.Sort = in.Sort
	if in.Status != nil {
		c.Status = *in.Status
	}
}

// CreateCourse 新建训练营；不传 status 默认草稿。
func (s *Service) CreateCourse(in CourseInput) (*AdminCourseItem, error) {
	c := &CoursePlan{Status: StatusDraft}
	applyCourseInput(c, in)
	if err := s.repo.CreateCourse(c); err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	item := toAdminCourseItem(c)
	return &item, nil
}

// ListCoursesAdmin 管理端分页列训练营（含草稿）。
func (s *Service) ListCoursesAdmin(status *int8, page, pageSize int) ([]AdminCourseItem, int64, int, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	courses, total, err := s.repo.ListCoursesAdmin(status, page, pageSize)
	if err != nil {
		return nil, 0, 0, 0, errcode.ErrServer.Wrap(err)
	}
	items := make([]AdminCourseItem, 0, len(courses))
	for i := range courses {
		items = append(items, toAdminCourseItem(&courses[i]))
	}
	return items, total, page, pageSize, nil
}

// UpdateCourse 编辑训练营。
func (s *Service) UpdateCourse(id uint, in CourseInput) (*AdminCourseItem, error) {
	c, err := s.repo.GetCourseByID(id)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if c == nil {
		return nil, errcode.ErrCourseNotFound
	}
	applyCourseInput(c, in)
	if err := s.repo.UpdateCourse(c); err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	item := toAdminCourseItem(c)
	return &item, nil
}

// UpdateCourseStatus 上下架训练营。
func (s *Service) UpdateCourseStatus(id uint, status int8) (*AdminCourseItem, error) {
	c, err := s.repo.GetCourseByID(id)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if c == nil {
		return nil, errcode.ErrCourseNotFound
	}
	c.Status = status
	if err := s.repo.UpdateCourse(c); err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	item := toAdminCourseItem(c)
	return &item, nil
}

// DeleteCourse 软删训练营。
func (s *Service) DeleteCourse(id uint) error {
	c, err := s.repo.GetCourseByID(id)
	if err != nil {
		return errcode.ErrServer.Wrap(err)
	}
	if c == nil {
		return errcode.ErrCourseNotFound
	}
	if err := s.repo.DeleteCourse(id); err != nil {
		return errcode.ErrServer.Wrap(err)
	}
	return nil
}

// buildCards 把入参词句卡转模型（lessonID 建/存时回填）。
func buildCards(in LessonInput) ([]LessonWord, []LessonPhrase) {
	words := make([]LessonWord, 0, len(in.Words))
	for _, w := range in.Words {
		words = append(words, LessonWord{
			Word: w.Word, Meaning: w.Meaning, Phonetic: w.Phonetic,
			ContextSentence: w.ContextSentence, IsRecur: w.IsRecur, RecurSeq: w.RecurSeq, Sort: w.Sort,
		})
	}
	phrases := make([]LessonPhrase, 0, len(in.Phrases))
	for _, p := range in.Phrases {
		phrases = append(phrases, LessonPhrase{Phrase: p.Phrase, Meaning: p.Meaning, Example: p.Example, Sort: p.Sort})
	}
	return words, phrases
}

// applyLessonInput 把入参写入课模型；type 为 0 默认工作日，status 仅传入时覆盖。
func applyLessonInput(l *DailyLesson, in LessonInput) {
	l.DayIndex = in.DayIndex
	l.Week = in.Week
	l.LessonType = in.LessonType
	if l.LessonType == 0 {
		l.LessonType = LessonTypeWeekday
	}
	l.VideoID = in.VideoID
	l.Title = in.Title
	l.Description = in.Description
	l.Sort = in.Sort
	if in.Status != nil {
		l.Status = *in.Status
	}
}

// adminLessonItem 组装管理端课项（含词句卡）。
func (s *Service) adminLessonItem(l *DailyLesson) (AdminLessonItem, error) {
	words, err := s.repo.ListWordsByLesson(l.ID)
	if err != nil {
		return AdminLessonItem{}, err
	}
	phrases, err := s.repo.ListPhrasesByLesson(l.ID)
	if err != nil {
		return AdminLessonItem{}, err
	}
	item := AdminLessonItem{
		ID: l.ID, CourseID: l.CourseID, DayIndex: l.DayIndex, Week: l.Week,
		LessonType: l.LessonType, VideoID: l.VideoID, Title: l.Title,
		Description: l.Description, Sort: l.Sort, Status: l.Status,
		Words: make([]LessonWordItem, 0, len(words)), Phrases: make([]LessonPhraseItem, 0, len(phrases)),
	}
	for i := range words {
		item.Words = append(item.Words, toWordItem(&words[i]))
	}
	for i := range phrases {
		item.Phrases = append(item.Phrases, toPhraseItem(&phrases[i]))
	}
	return item, nil
}

// CreateLesson 在某训练营下新建每日课（含词句卡）。
func (s *Service) CreateLesson(courseID uint, in LessonInput) (*AdminLessonItem, error) {
	c, err := s.repo.GetCourseByID(courseID)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if c == nil {
		return nil, errcode.ErrCourseNotFound
	}
	l := &DailyLesson{CourseID: courseID, Status: StatusDraft}
	applyLessonInput(l, in)
	words, phrases := buildCards(in)
	if err := s.repo.CreateLessonWithCards(l, words, phrases); err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	item, err := s.adminLessonItem(l)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	return &item, nil
}

// ListLessons 列某训练营全部课（含词句卡）。
func (s *Service) ListLessons(courseID uint) ([]AdminLessonItem, error) {
	c, err := s.repo.GetCourseByID(courseID)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if c == nil {
		return nil, errcode.ErrCourseNotFound
	}
	lessons, err := s.repo.ListLessonsByCourse(courseID)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	items := make([]AdminLessonItem, 0, len(lessons))
	for i := range lessons {
		item, err := s.adminLessonItem(&lessons[i])
		if err != nil {
			return nil, errcode.ErrServer.Wrap(err)
		}
		items = append(items, item)
	}
	return items, nil
}

// UpdateLesson 编辑每日课（词句卡整替换）。
func (s *Service) UpdateLesson(lessonID uint, in LessonInput) (*AdminLessonItem, error) {
	l, err := s.repo.GetLessonByID(lessonID)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if l == nil {
		return nil, errcode.ErrLessonNotFound
	}
	applyLessonInput(l, in)
	words, phrases := buildCards(in)
	if err := s.repo.UpdateLessonWithCards(l, words, phrases); err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	item, err := s.adminLessonItem(l)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	return &item, nil
}

// DeleteLesson 软删每日课。
func (s *Service) DeleteLesson(lessonID uint) error {
	l, err := s.repo.GetLessonByID(lessonID)
	if err != nil {
		return errcode.ErrServer.Wrap(err)
	}
	if l == nil {
		return errcode.ErrLessonNotFound
	}
	if err := s.repo.DeleteLesson(lessonID); err != nil {
		return errcode.ErrServer.Wrap(err)
	}
	return nil
}

// AdminEnroll 手动报名：校验用户+训练营存在→建报名（start_date 默认今天）。重复报名 ErrAlreadyEnrolled。
func (s *Service) AdminEnroll(req EnrollRequest) (*EnrollResponse, error) {
	exists, err := s.repo.UserExists(req.UserID)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if !exists {
		return nil, errcode.ErrUserNotFound
	}
	course, err := s.repo.GetCourseByID(req.CourseID)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if course == nil {
		return nil, errcode.ErrCourseNotFound
	}

	start := time.Now()
	if req.StartDate != "" {
		t, perr := time.ParseInLocation("2006-01-02", req.StartDate, time.Local)
		if perr != nil {
			return nil, errcode.ErrParam.WithMsg("start_date 格式应为 YYYY-MM-DD")
		}
		start = t
	}

	e := &Enrollment{UserID: req.UserID, CourseID: req.CourseID, StartDate: start, Status: EnrollmentStatusActive}
	if err := s.repo.CreateEnrollment(e); err != nil {
		if errors.Is(err, ErrDuplicateEnrollment) {
			return nil, errcode.ErrAlreadyEnrolled
		}
		return nil, errcode.ErrServer.Wrap(err)
	}
	return &EnrollResponse{
		EnrollmentID: e.ID, UserID: e.UserID, CourseID: e.CourseID,
		StartDate: e.StartDate.Format("2006-01-02"), Status: e.Status,
	}, nil
}
