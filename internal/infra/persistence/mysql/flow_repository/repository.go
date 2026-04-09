package flow_repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/google/uuid"

	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
)

// FlowRepositoryImpl GORM实现
type FlowRepositoryImpl struct {
	db *gorm.DB `autowired:""`
}

// Create 创建流程
func (r *FlowRepositoryImpl) Create(ctx context.Context, flow *entity.Flow) error {
	model, err := toModel(flow)
	if err != nil {
		return fmt.Errorf("convert to model: %w", err)
	}

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return fmt.Errorf("create flow: %w", result.Error)
	}

	// 同时创建版本记录
	versionModel, err := toVersionModel(flow)
	if err != nil {
		return fmt.Errorf("convert to version model: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(versionModel).Error; err != nil {
		return fmt.Errorf("create flow version: %w", err)
	}

	return nil
}

// Update 更新流程
func (r *FlowRepositoryImpl) Update(ctx context.Context, flow *entity.Flow) error {
	model, err := toModel(flow)
	if err != nil {
		return fmt.Errorf("convert to model: %w", err)
	}

	result := r.db.WithContext(ctx).Save(model)
	if result.Error != nil {
		return fmt.Errorf("update flow: %w", result.Error)
	}

	// 同时创建新版本记录
	versionModel, err := toVersionModel(flow)
	if err != nil {
		return fmt.Errorf("convert to version model: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(versionModel).Error; err != nil {
		return fmt.Errorf("create flow version: %w", err)
	}

	return nil
}

// Delete 删除流程
func (r *FlowRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	// 开始事务
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除主记录
		result := tx.Where("id = ?", id).Delete(&FlowModel{})
		if result.Error != nil {
			return fmt.Errorf("delete flow: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return domainErr.ErrNotFound
		}

		// 删除所有版本记录
		if err := tx.Where("flow_id = ?", id).Delete(&FlowVersionModel{}).Error; err != nil {
			return fmt.Errorf("delete flow versions: %w", err)
		}

		return nil
	})
}

// FindByID 根据ID查找流程
func (r *FlowRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.Flow, error) {
	var model FlowModel
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErr.ErrNotFound
		}
		return nil, fmt.Errorf("find flow by id: %w", result.Error)
	}

	return toEntity(&model)
}

// FindByWorkspaceID 根据工作空间ID查找流程
func (r *FlowRepositoryImpl) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Flow, error) {
	var models []FlowModel
	result := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Order("created_at DESC").Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("find flows by workspace id: %w", result.Error)
	}

	flows := make([]*entity.Flow, len(models))
	for i, model := range models {
		flow, err := toEntity(&model)
		if err != nil {
			return nil, fmt.Errorf("convert model %d to entity: %w", i, err)
		}
		flows[i] = flow
	}

	return flows, nil
}

// ListWithFilters 带过滤条件列出流程
func (r *FlowRepositoryImpl) ListWithFilters(ctx context.Context, workspaceID *string, status *entity.FlowStatus, offset, limit int) ([]*entity.Flow, int64, error) {
	db := r.db.WithContext(ctx).Model(&FlowModel{})

	if workspaceID != nil && *workspaceID != "" {
		wsID, err := uuid.Parse(*workspaceID)
		if err == nil {
			db = db.Where("workspace_id = ?", wsID)
		}
	}

	if status != nil {
		db = db.Where("status = ?", string(*status))
	}

	// 获取总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count flows: %w", err)
	}

	// 获取分页数据
	var models []FlowModel
	result := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&models)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("list flows with filters: %w", result.Error)
	}

	flows := make([]*entity.Flow, len(models))
	for i, model := range models {
		flow, err := toEntity(&model)
		if err != nil {
			return nil, 0, fmt.Errorf("convert model %d to entity: %w", i, err)
		}
		flows[i] = flow
	}

	return flows, total, nil
}

// FindByIDAndVersion 根据ID和版本查找流程
func (r *FlowRepositoryImpl) FindByIDAndVersion(ctx context.Context, id uuid.UUID, version int) (*entity.Flow, error) {
	var model FlowVersionModel
	result := r.db.WithContext(ctx).Where("flow_id = ? AND version = ?", id, version).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErr.ErrNotFound
		}
		return nil, fmt.Errorf("find flow by id and version: %w", result.Error)
	}

	return toEntityFromVersion(&model)
}

// SaveVersion 保存流程版本（用于版本管理）
func (r *FlowRepositoryImpl) SaveVersion(ctx context.Context, flow *entity.Flow) error {
	versionModel, err := toVersionModel(flow)
	if err != nil {
		return fmt.Errorf("convert to version model: %w", err)
	}

	result := r.db.WithContext(ctx).Create(versionModel)
	if result.Error != nil {
		return fmt.Errorf("save flow version: %w", result.Error)
	}

	return nil
}
