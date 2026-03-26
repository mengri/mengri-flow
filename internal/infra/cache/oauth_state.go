package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// OAuthStateStore OAuth state 参数 Redis 存储（CSRF 防护）。
type OAuthStateStore struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewOAuthStateStore 创建 OAuth state 存储。
func NewOAuthStateStore(rdb *redis.Client) *OAuthStateStore {
	return &OAuthStateStore{
		rdb: rdb,
		ttl: 5 * time.Minute,
	}
}

func oauthStateKey(state string) string {
	return fmt.Sprintf("oauth:state:%s", state)
}

// Generate 生成 state 并存入 Redis。
func (s *OAuthStateStore) Generate(ctx context.Context) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate oauth state: %w", err)
	}
	state := hex.EncodeToString(b)

	if err := s.rdb.Set(ctx, oauthStateKey(state), "1", s.ttl).Err(); err != nil {
		return "", fmt.Errorf("save oauth state: %w", err)
	}
	return state, nil
}

// Validate 验证并消费 state（一次性使用）。
func (s *OAuthStateStore) Validate(ctx context.Context, state string) error {
	key := oauthStateKey(state)
	result, err := s.rdb.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("validate oauth state: %w", err)
	}
	if result == 0 {
		return fmt.Errorf("oauth state invalid or expired")
	}
	return nil
}

// TTL 返回 state 有效期。
func (s *OAuthStateStore) TTL() time.Duration {
	return s.ttl
}
