package workspaceRepository

import (
	"context"
	"fmt"

	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/repository"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

// WorkspaceRepositoryImpl 是 WorkspaceRepository 的 GORM实现
type WorkspaceRepositoryImpl struct {
	db *gorm.DB `autowired:""`
}

var _ repository.WorkspaceRepository = (*WorkspaceRepositoryImpl)(nil)

// Create 创建工作空间
func (r *WorkspaceRepositoryImpl) Create(ctx context.Context, workspace *entity.Workspace) error {
	model := toModel(workspace)
	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return fmt.Errorf("workspaceRepository.Create: failed to create workspace: %w", result.Error)
	}
	workspace.ID = uuid.MustParse(model.ID)
	return nil
}

// Update 更新工作空间
func (r *WorkspaceRepositoryImpl) Update(ctx context.Context, workspace *entity.Workspace) error {
	model := toModel(workspace)
	result := r.db.WithContext(ctx).Model(&WorkspaceModel{}).Where("id = ?", model.ID).Updates(model)
	if result.Error != nil {
		return fmt.Errorf("workspaceRepository.Update: failed to update workspace: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domainErr.ErrNotFound
	}
	return nil
}

// Delete 删除工作空间
func (r *WorkspaceRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&WorkspaceModel{}, "id = ?", id.String())
	if result.Error != nil {
		return fmt.Errorf("workspaceRepository.Delete: failed to delete workspace: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domainErr.ErrNotFound
	}
	return nil
}

// FindByID 根据ID查找工作空间
func (r *WorkspaceRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.Workspace, error) {
	var model WorkspaceModel
	result := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErr.ErrNotFound
		}
		return nil, fmt.Errorf("workspaceRepository.FindByID: failed to find workspace: %w", result.Error)
	}
	return toEntity(&model)
}

// FindByOwnerID 根据所有者ID查找工作空间列表
func (r *WorkspaceRepositoryImpl) FindByOwnerID(ctx context.Context, ownerID string) ([]*entity.Workspace, error) {
	var models []WorkspaceModel
	result := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).
		Order("created_at DESC").Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("workspaceRepository.FindByOwnerID: failed to find workspaces: %w", result.Error)
	}

	workspaces := make([]*entity.Workspace, len(models))
	for i, model := range models {
		workspace, err := toEntity(&model)
		if err != nil {
			return nil, err
		}
		workspaces[i] = workspace
	}
	return workspaces, nil
}

// List 分页列出所有工作空间
func (r *WorkspaceRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*entity.Workspace, int64, error) {
	var models []WorkspaceModel
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&WorkspaceModel{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("workspaceRepository.List: failed to count workspaces: %w", err)
	}

	// 获取分页数据
	result := r.db.WithContext(ctx).Order("created_at DESC").
		Offset(offset).Limit(limit).Find(&models)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("workspaceRepository.List: failed to list workspaces: %w", result.Error)
	}

	workspaces := make([]*entity.Workspace, len(models))
	for i, model := range models {
		workspace, err := toEntity(&model)
		if err != nil {
			return nil, 0, err
		}
		workspaces[i] = workspace
	}
	return workspaces, total, nil
}

// ListByOwner 分页列出指定账号拥有的工作空间
func (r *WorkspaceRepositoryImpl) ListByOwner(ctx context.Context, ownerID string, offset, limit int) ([]*entity.Workspace, int64, error) {
	var models []WorkspaceModel
	var total int64

	query := r.db.WithContext(ctx).Model(&WorkspaceModel{}).Where("owner_id = ?", ownerID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("workspaceRepository.ListByOwner: failed to count workspaces: %w", err)
	}

	result := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).
		Order("created_at DESC").Offset(offset).Limit(limit).Find(&models)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("workspaceRepository.ListByOwner: failed to list workspaces: %w", result.Error)
	}

	workspaces := make([]*entity.Workspace, len(models))
	for i, model := range models {
		workspace, err := toEntity(&model)
		if err != nil {
			return nil, 0, err
		}
		workspaces[i] = workspace
	}
	return workspaces, total, nil
}
