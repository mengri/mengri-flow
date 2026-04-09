package repository

import (
	"context"
	"time"
)

// BindTicketData 绑定票据关联的数据。
type BindTicketData struct {
	Provider   string `json:"provider"`
	ExternalID string `json:"externalId"`
	Nickname   string `json:"nickname"`
	AvatarURL  string `json:"avatarUrl"`
}

// BindTicketStore 第三方绑定票据存储接口。
type BindTicketStore interface {
	Generate(ctx context.Context, data *BindTicketData) (string, error)
	Validate(ctx context.Context, ticket string) (*BindTicketData, error)
	TTL() time.Duration
}
