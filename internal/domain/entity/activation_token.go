package entity

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// ActivationToken 激活令牌实体。
type ActivationToken struct {
	TokenHash string
	AccountID string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

// NewActivationToken 创建激活令牌。rawToken 为明文 token，存储时使用 SHA-256 哈希。
func NewActivationToken(accountID string, rawToken string, ttl time.Duration) *ActivationToken {
	hash := sha256.Sum256([]byte(rawToken))
	now := time.Now()
	return &ActivationToken{
		TokenHash: hex.EncodeToString(hash[:]),
		AccountID: accountID,
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
	}
}

// HashToken 对明文 token 做 SHA-256 哈希。
func HashToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])
}

// IsValid 检查 token 是否有效（未过期 + 未使用）。
func (t *ActivationToken) IsValid() bool {
	return t.UsedAt == nil && time.Now().Before(t.ExpiresAt)
}

// IsExpired 检查 token 是否已过期。
func (t *ActivationToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed 检查 token 是否已使用。
func (t *ActivationToken) IsUsed() bool {
	return t.UsedAt != nil
}

// MarkUsed 标记已使用。
func (t *ActivationToken) MarkUsed() {
	now := time.Now()
	t.UsedAt = &now
}
