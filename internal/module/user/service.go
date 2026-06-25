package user

import (
	"context"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/echotalk/echotalk_server/internal/module/user/codesender"
	"github.com/echotalk/echotalk_server/internal/pkg/errcode"
	"github.com/echotalk/echotalk_server/internal/pkg/jwt"
)

// 验证码场景与有效期。
const (
	sceneRegister = "register"
	codeTTL       = 5 * time.Minute
)

// normalizeEmail 邮箱规范化：去空格 + 转小写，保证查重/验证码 key/登录一致。
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// Service 用户业务逻辑。
type Service struct {
	repo   *Repository
	jwt    *jwt.Manager
	sender codesender.CodeSender
	store  codesender.CodeStore
}

// NewService 创建服务。
func NewService(repo *Repository, jwtManager *jwt.Manager, sender codesender.CodeSender, store codesender.CodeStore) *Service {
	return &Service{repo: repo, jwt: jwtManager, sender: sender, store: store}
}

// SendCode 生成并送达验证码，落 Redis（带 TTL）。
func (s *Service) SendCode(ctx context.Context, req SendCodeRequest) error {
	email := normalizeEmail(req.Email)
	code, err := s.sender.Send(ctx, email)
	if err != nil {
		return errcode.ErrServer
	}
	if err := s.store.Save(ctx, sceneRegister, email, code, codeTTL); err != nil {
		return errcode.ErrServer
	}
	return nil
}

// Register 注册新用户。
// 顺序：规范化邮箱 → 查重(前置，不白烧验证码) → 校验验证码(一次性消费) → bcrypt → 写库。
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*User, error) {
	email := normalizeEmail(req.Email)

	existing, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, errcode.ErrServer
	}
	if existing != nil {
		return nil, errcode.ErrUserExists
	}

	ok, err := s.store.Consume(ctx, sceneRegister, email, req.Code)
	if err != nil {
		return nil, errcode.ErrServer
	}
	if !ok {
		return nil, errcode.ErrCodeInvalid
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errcode.ErrServer
	}
	u := &User{Email: email, Password: string(hash), Nickname: req.Nickname}
	if err := s.repo.Create(u); err != nil {
		if errors.Is(err, ErrDuplicateEmail) { // 并发竞态：唯一键冲突兜底
			return nil, errcode.ErrUserExists
		}
		return nil, errcode.ErrServer
	}
	return u, nil
}

// dummyHash 供登录时对不存在的用户也做一次等价 bcrypt 比对，抵御时序枚举。
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)

// Login 校验密码并签发令牌对。
// 防用户枚举：账号不存在与密码错误统一返回 ErrInvalidCredentials，且耗时一致。
func (s *Service) Login(req LoginRequest) (jwt.Pair, error) {
	u, err := s.repo.FindByEmail(normalizeEmail(req.Email))
	if err != nil {
		return jwt.Pair{}, errcode.ErrServer
	}
	if u == nil {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password)) // 等价耗时，结果丢弃
		return jwt.Pair{}, errcode.ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)) != nil {
		return jwt.Pair{}, errcode.ErrInvalidCredentials
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
