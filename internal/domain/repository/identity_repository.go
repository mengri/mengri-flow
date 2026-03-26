package repository

import (
	"context"
	"mengri-flow/internal/domain/entity"
)

// IdentityRepository 登录身份仓储接口。
type IdentityRepository interface {
	Create(ctx context.Context, identity *entity.Identity) error
	GetByID(ctx context.Context, id string) (*entity.Identity, error)
	GetByProviderID(ctx context.Context, loginType entity.LoginType, externalID string) (*entity.Identity, error)
	ListByAccountID(ctx context.Context, accountID string) ([]*entity.Identity, error)
	CountActiveByAccountID(ctx context.Context, accountID string) (int, error)
	SoftDelete(ctx context.Context, id string) error
}
