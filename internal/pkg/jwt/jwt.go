// Package jwt 封装 access + refresh 双令牌的签发与校验。
// refresh 令牌带 jti，配合服务端白名单实现可撤销（登出）。
package jwt

import (
	"crypto/rand"
	"encoding/hex"
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

// Claims 自定义声明。RegisteredClaims.ID 即 jti。
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

// RefreshTTL 返回 refresh 令牌有效期（供白名单设置同样的 TTL）。
func (m *Manager) RefreshTTL() time.Duration { return m.refreshTTL }

// GeneratePair 为用户签发一对令牌，并返回 refresh 令牌的 jti（用于服务端白名单）。
func (m *Manager) GeneratePair(userID uint) (Pair, string, error) {
	access, _, err := m.sign(userID, AccessToken, m.accessTTL)
	if err != nil {
		return Pair{}, "", err
	}
	refresh, refreshJTI, err := m.sign(userID, RefreshToken, m.refreshTTL)
	if err != nil {
		return Pair{}, "", err
	}
	return Pair{AccessToken: access, RefreshToken: refresh}, refreshJTI, nil
}

func (m *Manager) sign(userID uint, typ TokenType, ttl time.Duration) (string, string, error) {
	now := time.Now()
	jti, err := genJTI()
	if err != nil {
		return "", "", err
	}
	claims := Claims{
		UserID: userID,
		Type:   typ,
		RegisteredClaims: gojwt.RegisteredClaims{
			ID:        jti,
			Issuer:    m.issuer,
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(now.Add(ttl)),
		},
	}
	signed, err := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", "", err
	}
	return signed, jti, nil
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

func genJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
