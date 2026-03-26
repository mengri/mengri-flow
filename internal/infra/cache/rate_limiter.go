package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter 基于 Redis 的固定窗口速率限制器。
type RateLimiter struct {
	rdb *redis.Client
}

// NewRateLimiter 创建速率限制器。
func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{rdb: rdb}
}

// Allow 检查是否允许操作。
// key: 限制维度键（如 "login_fail:{accountId}"）
// maxAttempts: 窗口内最大次数
// window: 窗口时长
// 返回：是否允许、当前计数、error
func (l *RateLimiter) Allow(ctx context.Context, key string, maxAttempts int, window time.Duration) (bool, int, error) {
	fullKey := fmt.Sprintf("rate:%s", key)

	pipe := l.rdb.Pipeline()
	incrCmd := pipe.Incr(ctx, fullKey)
	pipe.Expire(ctx, fullKey, window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, fmt.Errorf("rate limiter: %w", err)
	}

	count := int(incrCmd.Val())
	return count <= maxAttempts, count, nil
}

// Reset 重置计数器。
func (l *RateLimiter) Reset(ctx context.Context, key string) error {
	fullKey := fmt.Sprintf("rate:%s", key)
	return l.rdb.Del(ctx, fullKey).Err()
}
