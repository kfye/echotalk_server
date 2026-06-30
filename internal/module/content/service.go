package content

import (
	"context"

	"github.com/echotalk/echotalk_server/internal/pkg/errcode"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

// Service 内容业务逻辑。
type Service struct {
	repo    *Repository
	members MembershipChecker
}

// NewService 创建服务。
func NewService(repo *Repository, members MembershipChecker) *Service {
	return &Service{repo: repo, members: members}
}

// locked 计算付费门禁：免费内容永不锁；付费内容仅会员可解锁。
func (s *Service) locked(ctx context.Context, userID uint, v *Video) (bool, error) {
	if v.IsFree {
		return false, nil
	}
	if userID == 0 { // 匿名
		return true, nil
	}
	ok, err := s.members.IsActiveMember(ctx, userID)
	if err != nil {
		return false, err
	}
	return !ok, nil
}

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

// List App 内容列表（仅上架），带付费锁标识。
func (s *Service) List(ctx context.Context, userID uint, q ListQuery) ([]VideoListItem, int64, int, int, error) {
	page, pageSize := normalizePage(q.Page, q.PageSize)
	videos, total, err := s.repo.List(ListFilter{
		OnlineOnly: true,
		Category:   q.Category,
		Difficulty: q.Difficulty,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		return nil, 0, 0, 0, errcode.ErrServer.Wrap(err)
	}
	items := make([]VideoListItem, 0, len(videos))
	for i := range videos {
		lk, err := s.locked(ctx, userID, &videos[i])
		if err != nil {
			return nil, 0, 0, 0, errcode.ErrServer.Wrap(err)
		}
		items = append(items, toListItem(&videos[i], lk))
	}
	return items, total, page, pageSize, nil
}

// Detail App 内容详情；付费未解锁时隐藏媒体地址。
func (s *Service) Detail(ctx context.Context, userID, id uint) (*VideoDetail, error) {
	v, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if v == nil || v.Status != VideoStatusOnline {
		return nil, errcode.ErrContentNotFound
	}
	lk, err := s.locked(ctx, userID, v)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	d := toDetail(v, lk)
	return &d, nil
}

// --- 管理端 ---

// AdminList 管理列表（含草稿/下架）。
func (s *Service) AdminList(q ListQuery, status *int8) ([]VideoDetail, int64, int, int, error) {
	page, pageSize := normalizePage(q.Page, q.PageSize)
	videos, total, err := s.repo.List(ListFilter{
		Status:   status,
		Category: q.Category,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, 0, 0, 0, errcode.ErrServer.Wrap(err)
	}
	items := make([]VideoDetail, 0, len(videos))
	for i := range videos {
		items = append(items, toDetail(&videos[i], false)) // 管理端不隐藏地址
	}
	return items, total, page, pageSize, nil
}

// Create 新增视频（默认草稿状态）。
func (s *Service) Create(in VideoInput) (*VideoDetail, error) {
	v := applyInput(&Video{Status: VideoStatusDraft, Difficulty: 1}, in)
	if err := s.repo.Create(v); err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	d := toDetail(v, false)
	return &d, nil
}

// Update 编辑视频。
func (s *Service) Update(id uint, in VideoInput) (*VideoDetail, error) {
	v, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if v == nil {
		return nil, errcode.ErrContentNotFound
	}
	v = applyInput(v, in)
	if err := s.repo.Update(v); err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	d := toDetail(v, false)
	return &d, nil
}

// UpdateStatus 上下架 / 免费付费切换。
func (s *Service) UpdateStatus(id uint, in UpdateStatusInput) (*VideoDetail, error) {
	v, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if v == nil {
		return nil, errcode.ErrContentNotFound
	}
	if in.Status != nil {
		v.Status = *in.Status
	}
	if in.IsFree != nil {
		v.IsFree = *in.IsFree
	}
	if err := s.repo.Update(v); err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	d := toDetail(v, false)
	return &d, nil
}

// Delete 删除（软删）。
func (s *Service) Delete(id uint) error {
	v, err := s.repo.GetByID(id)
	if err != nil {
		return errcode.ErrServer.Wrap(err)
	}
	if v == nil {
		return errcode.ErrContentNotFound
	}
	if err := s.repo.Delete(id); err != nil {
		return errcode.ErrServer.Wrap(err)
	}
	return nil
}

func applyInput(v *Video, in VideoInput) *Video {
	v.Title = in.Title
	v.Description = in.Description
	v.CoverURL = in.CoverURL
	v.HLSURL = in.HLSURL
	v.SubtitleEnURL = in.SubtitleEnURL
	v.SubtitleCnURL = in.SubtitleCnURL
	v.Duration = in.Duration
	if in.Difficulty > 0 {
		v.Difficulty = in.Difficulty
	}
	v.Category = in.Category
	v.IsFree = in.IsFree
	v.Sort = in.Sort
	return v
}
