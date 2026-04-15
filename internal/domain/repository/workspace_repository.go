package repository

import (
	"context"

	"mengri-flow/internal/domain/entity"

	"github.com/google/uuid"
)

// WorkspaceRepository 定义工作空间仓储接口
type WorkspaceRepository interface {
	// Create 创建工作空间
	Create(ctx context.Context, workspace *entity.Workspace) error

	// Update 更新工作空间
	Update(ctx context.Context, workspace *entity.Workspace) error

	// Delete 删除工作空间
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID 根据ID查找工作空间
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Workspace, error)

	// FindByOwnerID 根据所有者ID查找工作空间列表
	FindByOwnerID(ctx context.Context, ownerID string) ([]*entity.Workspace, error)

	// List 分页列出所有工作空间（管理用途）
	List(ctx context.Context, offset, limit int) ([]*entity.Workspace, int64, error)

	// ListByOwner 分页列出指定账号拥有的工作空间
	ListByOwner(ctx context.Context, ownerID string, offset, limit int) ([]*entity.Workspace, int64, error)
}
