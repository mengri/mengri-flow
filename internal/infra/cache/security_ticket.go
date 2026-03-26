package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"mengri-flow/pkg/autowire"
	"time"

	"github.com/redis/go-redis/v9"
)

// SecurityTicketStore 安全票据 Redis 存储。
type SecurityTicketStore struct {
	rdb *redis.Client
	ttl time.Duration
}

// GenSecurityTicketStore 创建安全票据存储。
func GenSecurityTicketStore(rdb *redis.Client, ttlSeconds int) {

	autowire.Auto(func() *SecurityTicketStore {
		return &SecurityTicketStore{
			rdb: rdb,
			ttl: time.Duration(ttlSeconds) * time.Second,
		}
	})

}

func securityTicketKey(ticket string) string {
	return fmt.Sprintf("security:ticket:%s", ticket)
}

// Generate 生成安全票据并存入 Redis。返回票据字符串。
func (s *SecurityTicketStore) Generate(ctx context.Context, accountID string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate security ticket: %w", err)
	}
	ticket := hex.EncodeToString(b)

	if err := s.rdb.Set(ctx, securityTicketKey(ticket), accountID, s.ttl).Err(); err != nil {
		return "", fmt.Errorf("save security ticket: %w", err)
	}
	return ticket, nil
}

// Validate 验证并消费安全票据（一次性使用）。返回 accountID。
func (s *SecurityTicketStore) Validate(ctx context.Context, ticket, expectedAccountID string) error {
	key := securityTicketKey(ticket)

	accountID, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("security ticket invalid or expired")
	}
	if err != nil {
		return fmt.Errorf("validate security ticket: %w", err)
	}

	if accountID != expectedAccountID {
		return fmt.Errorf("security ticket invalid or expired")
	}

	// 一次性使用，删除
	s.rdb.Del(ctx, key)
	return nil
}

// TTL 返回票据有效期。
func (s *SecurityTicketStore) TTL() time.Duration {
	return s.ttl
}
