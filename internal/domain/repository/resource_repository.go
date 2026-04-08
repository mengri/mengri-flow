package repository

import (
	"context"

	"github.com/google/uuid"
	"mengri-flow/internal/domain/entity"
)

// ResourceRepository 定义资源仓储接口
type ResourceRepository interface {
	// Create 创建资源
	Create(ctx context.Context, resource *entity.Resource) error

	// Update 更新资源
	Update(ctx context.Context, resource *entity.Resource) error

	// Delete 删除资源
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID 根据ID查找资源
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Resource, error)

	// ListByWorkspace 根据工作空间ID列出资源
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Resource, error)

	// ListByType 根据类型列出资源
	ListByType(ctx context.Context, resourceType entity.ResourceType) ([]*entity.Resource, error)

	// UpdateStatus 更新资源状态
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ResourceStatus) error
}
