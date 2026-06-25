package codesender

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCodeStore 基于 Redis 的验证码存储，key 形如 vcode:{scene}:{target}。
type RedisCodeStore struct {
	client *redis.Client
}

// NewRedisCodeStore 创建 Redis 存储。
func NewRedisCodeStore(client *redis.Client) *RedisCodeStore {
	return &RedisCodeStore{client: client}
}

// 编译期断言：确保实现了 CodeStore 接口。
var _ CodeStore = (*RedisCodeStore)(nil)

func key(scene, target string) string {
	return fmt.Sprintf("vcode:%s:%s", scene, target)
}

// Save 保存待验证码，TTL 后自动过期（覆盖同 key 旧码）。
func (s *RedisCodeStore) Save(ctx context.Context, scene, target, code string, ttl time.Duration) error {
	return s.client.Set(ctx, key(scene, target), code, ttl).Err()
}

// Consume 校验码：命中才删除（一次性消费）；错误码保留，可在 TTL 内重试。
func (s *RedisCodeStore) Consume(ctx context.Context, scene, target, code string) (bool, error) {
	k := key(scene, target)
	val, err := s.client.Get(ctx, k).Result()
	if err == redis.Nil {
		return false, nil // 不存在或已过期
	}
	if err != nil {
		return false, err
	}
	if val != code {
		return false, nil
	}
	// 命中：删除以保证一次性（并发严格场景可改用 Lua 原子 GET+DEL）。
	s.client.Del(ctx, k)
	return true, nil
}
