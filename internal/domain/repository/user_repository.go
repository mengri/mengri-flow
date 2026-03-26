package repository

import (
	"context"
	"mengri-flow/internal/domain/entity"
)

// UserRepository 定义用户仓储接口。
// 接口定义在 Domain 层（调用方），实现在 Infra 层（提供方）— 依赖倒置原则。
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uint64) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, offset, limit int) ([]*entity.User, int64, error)
}
