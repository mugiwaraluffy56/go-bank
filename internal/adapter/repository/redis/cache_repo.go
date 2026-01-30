package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yourusername/gobank/internal/domain/service"
	"github.com/yourusername/gobank/internal/infrastructure/database"
)

type cacheRepository struct {
	redis *database.RedisDB
}

func NewCacheRepository(redis *database.RedisDB) service.CacheService {
	return &cacheRepository{redis: redis}
}

func (r *cacheRepository) Get(ctx context.Context, key string) (string, error) {
	return r.redis.Get(ctx, key)
}

func (r *cacheRepository) Set(ctx context.Context, key string, value interface{}, ttlSeconds int) error {
	var data string
	switch v := value.(type) {
	case string:
		data = v
	default:
		bytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		data = string(bytes)
	}
	return r.redis.Set(ctx, key, data, time.Duration(ttlSeconds)*time.Second)
}

func (r *cacheRepository) Delete(ctx context.Context, key string) error {
	return r.redis.Delete(ctx, key)
}

func (r *cacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	return r.redis.Exists(ctx, key)
}

func (r *cacheRepository) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := r.redis.Get(ctx, key)
	if err != nil {
		return err
	}
	if data == "" {
		return nil
	}
	return json.Unmarshal([]byte(data), dest)
}

type RateLimiter struct {
	redis             *database.RedisDB
	requestsPerMinute int
	windowSize        time.Duration
}

func NewRateLimiter(redis *database.RedisDB, requestsPerMinute int) *RateLimiter {
	return &RateLimiter{
		redis:             redis,
		requestsPerMinute: requestsPerMinute,
		windowSize:        time.Minute,
	}
}

func (rl *RateLimiter) Allow(ctx context.Context, key string) (bool, int, error) {
	now := time.Now().Unix()
	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, now/60)

	count, err := rl.redis.Incr(ctx, windowKey)
	if err != nil {
		return false, 0, err
	}

	if count == 1 {
		if err := rl.redis.Expire(ctx, windowKey, rl.windowSize); err != nil {
			return false, 0, err
		}
	}

	remaining := rl.requestsPerMinute - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return count <= int64(rl.requestsPerMinute), remaining, nil
}

func (rl *RateLimiter) GetLimit() int {
	return rl.requestsPerMinute
}
