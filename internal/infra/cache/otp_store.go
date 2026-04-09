package cache

import (
	"context"
	"fmt"
	"mengri-flow/internal/domain/repository"
	"mengri-flow/pkg/autowire"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisOTPStore 是 OTPStore 的 Redis 实现。
type RedisOTPStore struct {
	rdb *redis.Client `autowired:""`
}

func init() {
	autowire.Auto(func() repository.OTPStore {
		return &RedisOTPStore{}
	})
}

var _ repository.OTPStore = (*RedisOTPStore)(nil)

func otpKey(scene, target string) string {
	return fmt.Sprintf("otp:%s:%s", scene, target)
}

func otpRateKey(target string) string {
	return fmt.Sprintf("otp:rate:%s", target)
}

// Save 保存 OTP 哈希值，设置 TTL。
func (s *RedisOTPStore) Save(ctx context.Context, scene, target, codeHash string, ttl time.Duration) error {
	return s.rdb.Set(ctx, otpKey(scene, target), codeHash, ttl).Err()
}

// Get 获取 OTP 哈希值。
func (s *RedisOTPStore) Get(ctx context.Context, scene, target string) (string, error) {
	val, err := s.rdb.Get(ctx, otpKey(scene, target)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// Delete 删除 OTP。
func (s *RedisOTPStore) Delete(ctx context.Context, scene, target string) error {
	return s.rdb.Del(ctx, otpKey(scene, target)).Err()
}

// IncrSendCount 递增发送计数器，用于频率限制。
func (s *RedisOTPStore) IncrSendCount(ctx context.Context, target string, window time.Duration) (int, error) {
	key := otpRateKey(target)
	pipe := s.rdb.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return int(incrCmd.Val()), nil
}
