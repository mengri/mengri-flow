package repository

import (
	"context"
	"mengri-flow/internal/domain/entity"
)

// ActivationTokenRepository 激活令牌仓储接口。
type ActivationTokenRepository interface {
	Create(ctx context.Context, token *entity.ActivationToken) error
	GetByHash(ctx context.Context, tokenHash string) (*entity.ActivationToken, error)
	InvalidateByAccountID(ctx context.Context, accountID string) error
	MarkUsed(ctx context.Context, tokenHash string) error
}
