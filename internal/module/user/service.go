package user

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/echotalk/echotalk_server/internal/module/user/codesender"
	"github.com/echotalk/echotalk_server/internal/pkg/errcode"
	"github.com/echotalk/echotalk_server/internal/pkg/jwt"
)

// Service 用户业务逻辑。
type Service struct {
	repo   *Repository
	jwt    *jwt.Manager
	sender codesender.CodeSender
}

// NewService 创建服务。
func NewService(repo *Repository, jwtManager *jwt.Manager, sender codesender.CodeSender) *Service {
	return &Service{repo: repo, jwt: jwtManager, sender: sender}
}

// SendCode 发送注册验证码。
func (s *Service) SendCode(ctx context.Context, req SendCodeRequest) error {
	return s.sender.Send(ctx, req.Email)
}

// Register 注册新用户。
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*User, error) {
	if !s.sender.Verify(ctx, req.Email, req.Code) {
		return nil, errcode.ErrCodeInvalid
	}
	existing, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return nil, errcode.ErrServer
	}
	if existing != nil {
		return nil, errcode.ErrUserExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errcode.ErrServer
	}
	u := &User{Email: req.Email, Password: string(hash), Nickname: req.Nickname}
	if err := s.repo.Create(u); err != nil {
		return nil, errcode.ErrServer
	}
	return u, nil
}

// Login 校验密码并签发令牌对。
func (s *Service) Login(req LoginRequest) (jwt.Pair, error) {
	u, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return jwt.Pair{}, errcode.ErrServer
	}
	if u == nil {
		return jwt.Pair{}, errcode.ErrUserNotFound
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)) != nil {
		return jwt.Pair{}, errcode.ErrPasswordWrong
	}
	pair, err := s.jwt.GeneratePair(u.ID)
	if err != nil {
		return jwt.Pair{}, errcode.ErrServer
	}
	return pair, nil
}

// Refresh 用 refresh 令牌换新令牌对。
func (s *Service) Refresh(req RefreshRequest) (jwt.Pair, error) {
	pair, err := s.jwt.Refresh(req.RefreshToken)
	if err != nil {
		return jwt.Pair{}, errcode.ErrTokenInvalid
	}
	return pair, nil
}

// Profile 查询个人信息。
func (s *Service) Profile(userID uint) (*User, error) {
	u, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, errcode.ErrServer
	}
	if u == nil {
		return nil, errcode.ErrUserNotFound
	}
	return u, nil
}
