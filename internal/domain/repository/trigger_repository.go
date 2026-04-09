package repository

import (
	"context"

	"github.com/google/uuid"
	"mengri-flow/internal/domain/entity"
)

// TriggerRepository 定义触发器仓储接口
type TriggerRepository interface {
	// Create 创建触发器
	Create(ctx context.Context, trigger *entity.Trigger) error

	// Update 更新触发器
	Update(ctx context.Context, trigger *entity.Trigger) error

	// Delete 删除触发器
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID 根据ID查找触发器
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Trigger, error)

	// FindByFlowID 根据FlowID查找触发器列表
	FindByFlowID(ctx context.Context, flowID uuid.UUID) ([]*entity.Trigger, error)

	// ListWithFilters 根据条件列出触发器
	ListWithFilters(ctx context.Context, flowID *string, status *entity.TriggerStatus, offset, limit int) ([]*entity.Trigger, int64, error)
}