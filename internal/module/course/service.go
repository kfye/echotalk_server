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
