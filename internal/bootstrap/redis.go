package bootstrap

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/echotalk/echotalk_server/internal/config"
)

// InitRedis 初始化并探活 Redis 连接。
func InitRedis(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
