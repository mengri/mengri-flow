package repository

import (
	"context"

	"github.com/google/uuid"
	"mengri-flow/internal/domain/entity"
)

// ToolRepository 定义工具仓储接口
type ToolRepository interface {
	// Create 创建工具
	Create(ctx context.Context, tool *entity.Tool) error

	// Update 更新工具
	Update(ctx context.Context, tool *entity.Tool) error

	// Delete 删除工具
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID 根据ID查找工具
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Tool, error)

	// FindByWorkspaceID 根据WorkspaceID查找工具列表
	FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Tool, error)

	// ListWithFilters 根据条件列出工具
	ListWithFilters(ctx context.Context, workspaceID *string, toolType *string, status *entity.ToolStatus, offset, limit int) ([]*entity.Tool, int64, error)
}