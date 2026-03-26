package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// BindTicketData 绑定票据关联的数据。
type BindTicketData struct {
	Provider   string `json:"provider"`
	ExternalID string `json:"externalId"`
	Nickname   string `json:"nickname"`
	AvatarURL  string `json:"avatarUrl"`
}

// BindTicketStore 第三方绑定票据 Redis 存储。
type BindTicketStore struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewBindTicketStore 创建绑定票据存储。
func NewBindTicketStore(rdb *redis.Client) *BindTicketStore {
	return &BindTicketStore{
		rdb: rdb,
		ttl: 5 * time.Minute,
	}
}

func bindTicketKey(ticket string) string {
	return fmt.Sprintf("bind:ticket:%s", ticket)
}

// Generate 生成绑定票据并存入 Redis。
func (s *BindTicketStore) Generate(ctx context.Context, data *BindTicketData) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate bind ticket: %w", err)
	}
	ticket := hex.EncodeToString(b)

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal bind ticket data: %w", err)
	}

	if err := s.rdb.Set(ctx, bindTicketKey(ticket), string(jsonBytes), s.ttl).Err(); err != nil {
		return "", fmt.Errorf("save bind ticket: %w", err)
	}
	return ticket, nil
}

// Validate 验证并消费绑定票据（一次性使用），返回关联数据。
func (s *BindTicketStore) Validate(ctx context.Context, ticket string) (*BindTicketData, error) {
	key := bindTicketKey(ticket)

	val, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("bind ticket invalid or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("validate bind ticket: %w", err)
	}

	// 一次性使用
	s.rdb.Del(ctx, key)

	var data BindTicketData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("unmarshal bind ticket data: %w", err)
	}
	return &data, nil
}

// TTL 返回票据有效期。
func (s *BindTicketStore) TTL() time.Duration {
	return s.ttl
}
