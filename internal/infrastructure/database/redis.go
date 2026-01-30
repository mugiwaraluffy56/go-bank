package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yourusername/gobank/internal/infrastructure/config"
)

type RedisDB struct {
	Client *redis.Client
}

func NewRedisDB(ctx context.Context, cfg *config.RedisConfig) (*RedisDB, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisDB{Client: client}, nil
}

func (r *RedisDB) Close() error {
	if r.Client != nil {
		return r.Client.Close()
	}
	return nil
}

func (r *RedisDB) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

func (r *RedisDB) Get(ctx context.Context, key string) (string, error) {
	val, err := r.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (r *RedisDB) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisDB) Delete(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

func (r *RedisDB) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.Client.Exists(ctx, key).Result()
	return n > 0, err
}

func (r *RedisDB) Incr(ctx context.Context, key string) (int64, error) {
	return r.Client.Incr(ctx, key).Result()
}

func (r *RedisDB) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.Client.Expire(ctx, key, expiration).Err()
}

func (r *RedisDB) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.Client.TTL(ctx, key).Result()
}
