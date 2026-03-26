package repository

import (
	"context"
	"mengri-flow/internal/domain/entity"
)

// AccountRepository 账号仓储接口。
type AccountRepository interface {
	Create(ctx context.Context, account *entity.Account) error
	GetByID(ctx context.Context, id string) (*entity.Account, error)
	GetByEmail(ctx context.Context, email string) (*entity.Account, error)
	GetByUsername(ctx context.Context, username string) (*entity.Account, error)
	Update(ctx context.Context, account *entity.Account) error
	List(ctx context.Context, offset, limit int, status *entity.AccountStatus, keyword string) ([]*entity.Account, int64, error)
}
