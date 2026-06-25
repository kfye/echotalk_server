// Package jwt 封装 access + refresh 双令牌的签发与校验。
package jwt

import (
	"errors"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

// TokenType 令牌类型。
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims 自定义声明。
type Claims struct {
	UserID uint      `json:"uid"`
	Type   TokenType `json:"typ"`
	gojwt.RegisteredClaims
}

// Pair access + refresh 令牌对。
type Pair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Manager 令牌管理器。
type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	issuer     string
}

// NewManager 创建令牌管理器。
func NewManager(secret string, accessTTL, refreshTTL time.Duration, issuer string) *Manager {
	return &Manager{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL, issuer: issuer}
}

// GeneratePair 为用户签发一对令牌。
func (m *Manager) GeneratePair(userID uint) (Pair, error) {
	access, err := m.sign(userID, AccessToken, m.accessTTL)
	if err != nil {
		return Pair{}, err
	}
	refresh, err := m.sign(userID, RefreshToken, m.refreshTTL)
	if err != nil {
		return Pair{}, err
	}
	return Pair{AccessToken: access, RefreshToken: refresh}, nil
}

func (m *Manager) sign(userID uint, typ TokenType, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Type:   typ,
		RegisteredClaims: gojwt.RegisteredClaims{
			Issuer:    m.issuer,
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse 校验并解析令牌。
func (m *Manager) Parse(token string) (*Claims, error) {
	parsed, err := gojwt.ParseWithClaims(token, &Claims{}, func(t *gojwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// Refresh 用 refresh 令牌换取新的令牌对。
func (m *Manager) Refresh(refreshToken string) (Pair, error) {
	claims, err := m.Parse(refreshToken)
	if err != nil {
		return Pair{}, err
	}
	if claims.Type != RefreshToken {
		return Pair{}, errors.New("not a refresh token")
	}
	return m.GeneratePair(claims.UserID)
}
