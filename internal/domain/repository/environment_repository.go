package repository

import (
	"context"

	"github.com/google/uuid"
	"mengri-flow/internal/domain/entity"
)

// EnvironmentRepository 定义环境仓储接口
type EnvironmentRepository interface {
	// Create 创建环境
	Create(ctx context.Context, env *entity.Environment) error

	// Update 更新环境
	Update(ctx context.Context, env *entity.Environment) error

	// Delete 删除环境
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID 根据ID查找环境
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Environment, error)

	// FindByKey 根据Key查找环境
	FindByKey(ctx context.Context, key string) (*entity.Environment, error)

	// List 列出所有环境
	List(ctx context.Context) ([]*entity.Environment, error)
}
