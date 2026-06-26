// Package tokenstore 维护 refresh 令牌白名单，支持登出撤销（全部会话）。
// 纯无状态 JWT 无法登出；这里把 refresh 的 jti 存 Redis，登出即删除。
package tokenstore

import (
	"context"
	"time"
)

// TokenStore refresh 令牌白名单存储。
type TokenStore interface {
	// SaveRefresh 把某用户的一个 refresh jti 加入白名单，TTL=refresh 有效期。
	SaveRefresh(ctx context.Context, userID uint, jti string, ttl time.Duration) error
	// IsRefreshValid 判断该 jti 是否仍在白名单（未被撤销）。
	IsRefreshValid(ctx context.Context, userID uint, jti string) (bool, error)
	// RevokeRefresh 撤销单个 jti（刷新轮换时删旧码）。
	RevokeRefresh(ctx context.Context, userID uint, jti string) error
	// RevokeAllRefresh 撤销该用户全部 refresh 会话（登出/踢下线）。
	RevokeAllRefresh(ctx context.Context, userID uint) error
}
