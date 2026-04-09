package repository

import (
	"context"

	"github.com/google/uuid"

	"mengri-flow/internal/domain/entity"
)

// FlowRepository 流程仓库接口
type FlowRepository interface {
	// Create 创建流程
	Create(ctx context.Context, flow *entity.Flow) error
	// Update 更新流程
	Update(ctx context.Context, flow *entity.Flow) error
	// Delete 删除流程
	Delete(ctx context.Context, id uuid.UUID) error
	// FindByID 根据ID查找流程
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Flow, error)
	// FindByWorkspaceID 根据工作空间ID查找流程
	FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Flow, error)
	// ListWithFilters 带过滤条件列出流程
	ListWithFilters(ctx context.Context, workspaceID *string, status *entity.FlowStatus, offset, limit int) ([]*entity.Flow, int64, error)
	// FindByIDAndVersion 根据ID和版本查找流程（用于版本管理）
	FindByIDAndVersion(ctx context.Context, id uuid.UUID, version int) (*entity.Flow, error)
	// SaveVersion 保存流程版本（用于版本管理）
	SaveVersion(ctx context.Context, flow *entity.Flow) error
}