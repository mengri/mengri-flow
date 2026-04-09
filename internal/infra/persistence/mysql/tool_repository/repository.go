package tool_repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/google/uuid"

	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
)

// ToolRepositoryImpl GORM实现
type ToolRepositoryImpl struct {
	db *gorm.DB `autowired:""`
}

// Create 创建工具
func (r *ToolRepositoryImpl) Create(ctx context.Context, tool *entity.Tool) error {
	model, err := toModel(tool)
	if err != nil {
		return fmt.Errorf("convert to model: %w", err)
	}

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return fmt.Errorf("create tool: %w", result.Error)
	}

	return nil
}

// Update 更新工具
func (r *ToolRepositoryImpl) Update(ctx context.Context, tool *entity.Tool) error {
	model, err := toModel(tool)
	if err != nil {
		return fmt.Errorf("convert to model: %w", err)
	}

	result := r.db.WithContext(ctx).Save(model)
	if result.Error != nil {
		return fmt.Errorf("update tool: %w", result.Error)
	}

	return nil
}

// Delete 删除工具
func (r *ToolRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&ToolModel{})
	if result.Error != nil {
		return fmt.Errorf("delete tool: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domainErr.ErrNotFound
	}

	return nil
}

// FindByID 根据ID查找工具
func (r *ToolRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.Tool, error) {
	var model ToolModel
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErr.ErrNotFound
		}
		return nil, fmt.Errorf("find tool by id: %w", result.Error)
	}

	return toEntity(&model)
}

// FindByWorkspaceID 根据工作空间ID查找工具
func (r *ToolRepositoryImpl) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Tool, error) {
	var models []ToolModel
	result := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Order("created_at DESC").Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("find tools by workspace id: %w", result.Error)
	}

	tools := make([]*entity.Tool, len(models))
	for i, model := range models {
		tool, err := toEntity(&model)
		if err != nil {
			return nil, fmt.Errorf("convert model %d to entity: %w", i, err)
		}
		tools[i] = tool
	}

	return tools, nil
}

// ListWithFilters 带过滤条件列出工具
func (r *ToolRepositoryImpl) ListWithFilters(ctx context.Context, workspaceID *string, toolType *string, status *entity.ToolStatus, offset, limit int) ([]*entity.Tool, int64, error) {
	db := r.db.WithContext(ctx).Model(&ToolModel{})

	if workspaceID != nil && *workspaceID != "" {
		wsID, err := uuid.Parse(*workspaceID)
		if err == nil {
			db = db.Where("workspace_id = ?", wsID)
		}
	}

	if toolType != nil && *toolType != "" {
		db = db.Where("type = ?", *toolType)
	}

	if status != nil {
		db = db.Where("status = ?", string(*status))
	}

	// 获取总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count tools: %w", err)
	}

	// 获取分页数据
	var models []ToolModel
	result := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&models)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("list tools with filters: %w", result.Error)
	}

	tools := make([]*entity.Tool, len(models))
	for i, model := range models {
		tool, err := toEntity(&model)
		if err != nil {
			return nil, 0, fmt.Errorf("convert model %d to entity: %w", i, err)
		}
		tools[i] = tool
	}

	return tools, total, nil
}
