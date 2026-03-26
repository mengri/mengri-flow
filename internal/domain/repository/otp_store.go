package repository

import (
	"context"
	"time"
)

// OTPStore 验证码存储接口（Redis 实现）。
type OTPStore interface {
	Save(ctx context.Context, scene, target, codeHash string, ttl time.Duration) error
	Get(ctx context.Context, scene, target string) (codeHash string, err error)
	Delete(ctx context.Context, scene, target string) error
	IncrSendCount(ctx context.Context, target string, window time.Duration) (count int, err error)
}
