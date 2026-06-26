package tokenstore

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisTokenStore 用 Redis Set 维护每个用户的活跃 refresh jti 集合（key=urt:{userID}）。
type RedisTokenStore struct {
	client *redis.Client
}

// NewRedisTokenStore 创建 Redis 令牌存储。
func NewRedisTokenStore(client *redis.Client) *RedisTokenStore {
	return &RedisTokenStore{client: client}
}

// 编译期断言：确保实现了 TokenStore 接口。
var _ TokenStore = (*RedisTokenStore)(nil)

func userKey(userID uint) string {
	return fmt.Sprintf("urt:%d", userID)
}

// SaveRefresh 加入白名单并刷新整集合的过期时间。
func (s *RedisTokenStore) SaveRefresh(ctx context.Context, userID uint, jti string, ttl time.Duration) error {
	k := userKey(userID)
	if err := s.client.SAdd(ctx, k, jti).Err(); err != nil {
		return err
	}
	return s.client.Expire(ctx, k, ttl).Err()
}

// IsRefreshValid 判断 jti 是否在白名单。
func (s *RedisTokenStore) IsRefreshValid(ctx context.Context, userID uint, jti string) (bool, error) {
	return s.client.SIsMember(ctx, userKey(userID), jti).Result()
}

// RevokeRefresh 移除单个 jti。
func (s *RedisTokenStore) RevokeRefresh(ctx context.Context, userID uint, jti string) error {
	return s.client.SRem(ctx, userKey(userID), jti).Err()
}

// RevokeAllRefresh 删除整个集合 = 撤销该用户全部 refresh 会话。
func (s *RedisTokenStore) RevokeAllRefresh(ctx context.Context, userID uint) error {
	return s.client.Del(ctx, userKey(userID)).Err()
}
