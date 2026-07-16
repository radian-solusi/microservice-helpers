package connections

import (
	"context"
	"fmt"
	"strconv"
	"time"

	helperconfig "github.com/radian-solusi/go-helpers/config"
	"github.com/redis/go-redis/v9"
)

type redisWrapper struct {
	client *redis.Client
}

func NewRedis(cfg helperconfig.RedisConfig) Redis {
	client := redis.NewClient(&redis.Options{
		Addr:             cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Password:         cfg.Password,
		DB:               cfg.DB,
		DisableIndentity: true,
		Protocol:         2,
		DialTimeout:      5 * time.Second,
		ReadTimeout:      3 * time.Second,
		WriteTimeout:     3 * time.Second,
		PoolSize:         10,
		MinIdleConns:     2,
		MaxRetries:       2,
	})
	return &redisWrapper{client: client}
}

func (r *redisWrapper) Client() *redis.Client { return r.client }

func (r *redisWrapper) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *redisWrapper) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	if err := r.client.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("redis set %q: %w", key, err)
	}
	return nil
}

func (r *redisWrapper) Get(ctx context.Context, key string) (string, error) {
	res, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("redis get %q: %w", key, err)
	}
	return res, nil
}

func (r *redisWrapper) Clear(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis delete %q: %w", key, err)
	}
	return nil
}

func (r *redisWrapper) ClearPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	var deletedCount int
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("redis scan %q: %w", pattern, err)
		}
		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("redis delete keys: %w", err)
			}
			deletedCount += len(keys)
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

func (r *redisWrapper) Close() error {
	return r.client.Close()
}
