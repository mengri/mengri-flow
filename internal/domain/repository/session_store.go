package repository

import (
	"context"
	"time"
)

// SessionStore 会话存储接口（RefreshToken 管理）。
type SessionStore interface {
	SaveRefreshToken(ctx context.Context, sessionID, accountID, tokenHash, deviceInfoJSON, ip string, ttl time.Duration) error
	ValidateRefreshToken(ctx context.Context, tokenHash string) (accountID string, err error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllByAccountID(ctx context.Context, accountID string, exceptTokenHash string) (count int, err error)
}
