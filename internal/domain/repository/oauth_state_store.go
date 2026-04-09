package repository

import (
	"context"
	"time"
)

// OAuthStateStore OAuth state 参数存储接口（CSRF 防护）。
type OAuthStateStore interface {
	Generate(ctx context.Context) (string, error)
	Validate(ctx context.Context, state string) error
	TTL() time.Duration
}
