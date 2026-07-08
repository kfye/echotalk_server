package user

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/echotalk/echotalk_server/internal/module/user/codesender"
	"github.com/echotalk/echotalk_server/internal/module/user/tokenstore"
	"github.com/echotalk/echotalk_server/internal/pkg/errcode"
	"github.com/echotalk/echotalk_server/internal/pkg/jwt"
)

// 验证码场景与有效期。
const (
	sceneRegister = "register"
	codeTTL       = 5 * time.Minute
)

// phoneRe 中国大陆手机号：1 开头、第二位 3-9、共 11 位。
var phoneRe = regexp.MustCompile(`^1[3-9]\d{9}$`)

// normalizePhone 手机号规范化：去首尾空格（手机号无大小写），保证查重/验证码 key/登录一致。
func normalizePhone(phone string) string {
	return strings.TrimSpace(phone)
}

// Service 用户业务逻辑。
type Service struct {
	repo   *Repository
	jwt    *jwt.Manager
	sender codesender.CodeSender
	store  codesender.CodeStore
	tokens tokenstore.TokenStore
}

// NewService 创建服务。
func NewService(repo *Repository, jwtManager *jwt.Manager, sender codesender.CodeSender, store codesender.CodeStore, tokens tokenstore.TokenStore) *Service {
	return &Service{repo: repo, jwt: jwtManager, sender: sender, store: store, tokens: tokens}
}

// issuePair 签发令牌对，并把 refresh jti 加入白名单。
func (s *Service) issuePair(ctx context.Context, userID uint) (jwt.Pair, error) {
	pair, refreshJTI, err := s.jwt.GeneratePair(userID)
	if err != nil {
		return jwt.Pair{}, errcode.ErrServer.Wrap(err)
	}
	if err := s.tokens.SaveRefresh(ctx, userID, refreshJTI, s.jwt.RefreshTTL()); err != nil {
		return jwt.Pair{}, errcode.ErrServer.Wrap(err)
	}
	return pair, nil
}

// SendCode 生成并送达验证码，落 Redis（带 TTL）。
func (s *Service) SendCode(ctx context.Context, req SendCodeRequest) error {
	phone := normalizePhone(req.Phone)
	if !phoneRe.MatchString(phone) {
		return errcode.ErrParam.WithMsg("手机号格式不正确")
	}
	code, err := s.sender.Send(ctx, phone)
	if err != nil {
		return errcode.ErrServer.Wrap(err)
	}
	if err := s.store.Save(ctx, sceneRegister, phone, code, codeTTL); err != nil {
		return errcode.ErrServer.Wrap(err)
	}
	return nil
}

// Register 注册新用户。
// 顺序：规范化手机号 → 格式校验 → 查重(前置，不白烧验证码) → 校验验证码(一次性消费) → bcrypt → 写库。
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*User, error) {
	phone := normalizePhone(req.Phone)
	if !phoneRe.MatchString(phone) {
		return nil, errcode.ErrParam.WithMsg("手机号格式不正确")
	}

	existing, err := s.repo.FindByPhone(phone)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if existing != nil {
		return nil, errcode.ErrUserExists
	}

	ok, err := s.store.Consume(ctx, sceneRegister, phone, req.Code)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if !ok {
		return nil, errcode.ErrCodeInvalid
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	u := &User{Phone: phone, Password: string(hash), Nickname: req.Nickname}
	if err := s.repo.Create(u); err != nil {
		if errors.Is(err, ErrDuplicatePhone) { // 并发竞态：唯一键冲突兜底
			return nil, errcode.ErrUserExists
		}
		return nil, errcode.ErrServer.Wrap(err)
	}
	return u, nil
}

// dummyHash 供登录时对不存在的用户也做一次等价 bcrypt 比对，抵御时序枚举。
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)

// Login 校验密码并签发令牌对。
// 防用户枚举：账号不存在与密码错误统一返回 ErrInvalidCredentials，且耗时一致。
func (s *Service) Login(ctx context.Context, req LoginRequest) (jwt.Pair, error) {
	u, err := s.repo.FindByPhone(normalizePhone(req.Phone))
	if err != nil {
		return jwt.Pair{}, errcode.ErrServer.Wrap(err)
	}
	if u == nil {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password)) // 等价耗时，结果丢弃
		return jwt.Pair{}, errcode.ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)) != nil {
		return jwt.Pair{}, errcode.ErrInvalidCredentials
	}
	return s.issuePair(ctx, u.ID)
}

// Refresh 用 refresh 令牌换新令牌对：校验白名单 + 轮换（删旧 jti、存新 jti）。
func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (jwt.Pair, error) {
	claims, err := s.jwt.Parse(req.RefreshToken)
	if err != nil || claims.Type != jwt.RefreshToken {
		return jwt.Pair{}, errcode.ErrTokenInvalid
	}
	ok, err := s.tokens.IsRefreshValid(ctx, claims.UserID, claims.ID)
	if err != nil {
		return jwt.Pair{}, errcode.ErrServer.Wrap(err)
	}
	if !ok { // 已登出或已被轮换
		return jwt.Pair{}, errcode.ErrTokenInvalid
	}
	pair, err := s.issuePair(ctx, claims.UserID)
	if err != nil {
		return jwt.Pair{}, err
	}
	_ = s.tokens.RevokeRefresh(ctx, claims.UserID, claims.ID) // 旧 refresh 作废，防重放
	return pair, nil
}

// Logout 撤销该用户全部 refresh 会话（多端下线）。
func (s *Service) Logout(ctx context.Context, userID uint) error {
	if err := s.tokens.RevokeAllRefresh(ctx, userID); err != nil {
		return errcode.ErrServer.Wrap(err)
	}
	return nil
}

// Profile 查询个人信息。
func (s *Service) Profile(userID uint) (*User, error) {
	u, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, errcode.ErrServer.Wrap(err)
	}
	if u == nil {
		return nil, errcode.ErrUserNotFound
	}
	return u, nil
}
