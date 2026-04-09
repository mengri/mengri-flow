package resourceRepository

import (
	"context"
	"fmt"
	"time"

	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/repository"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

// ResourceRepositoryImpl 是 ResourceRepository 的 GORM实现
type ResourceRepositoryImpl struct {
	db *gorm.DB `autowired:""`
}

var _ repository.ResourceRepository = (*ResourceRepositoryImpl)(nil)

// Create 创建资源
func (r *ResourceRepositoryImpl) Create(ctx context.Context, resource *entity.Resource) error {
	model := toModel(resource)
	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return fmt.Errorf("resourceRepository.Create: failed to create resource: %w", result.Error)
	}
	resource.ID = uuid.MustParse(model.ID)
	return nil
}

// Update 更新资源
func (r *ResourceRepositoryImpl) Update(ctx context.Context, resource *entity.Resource) error {
	model := toModel(resource)
	result := r.db.WithContext(ctx).Model(&ResourceModel{}).Where("id = ?", model.ID).Updates(model)
	if result.Error != nil {
		return fmt.Errorf("resourceRepository.Update: failed to update resource: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domainErr.ErrNotFound
	}
	return nil
}

// Delete 删除资源
func (r *ResourceRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&ResourceModel{}, "id = ?", id.String())
	if result.Error != nil {
		return fmt.Errorf("resourceRepository.Delete: failed to delete resource: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domainErr.ErrNotFound
	}
	return nil
}

// FindByID 根据ID查找资源
func (r *ResourceRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.Resource, error) {
	var model ResourceModel
	result := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErr.ErrNotFound
		}
		return nil, fmt.Errorf("resourceRepository.FindByID: failed to find resource: %w", result.Error)
	}
	return toEntity(&model)
}

// ListByWorkspace 根据工作空间ID列出资源
func (r *ResourceRepositoryImpl) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Resource, error) {
	var models []ResourceModel
	result := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID.String()).
		Order("created_at DESC").Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("resourceRepository.ListByWorkspace: failed to list resources: %w", result.Error)
	}

	resources := make([]*entity.Resource, len(models))
	for i, model := range models {
		resource, err := toEntity(&model)
		if err != nil {
			return nil, err
		}
		resources[i] = resource
	}
	return resources, nil
}

// ListByType 根据类型列出资源
func (r *ResourceRepositoryImpl) ListByType(ctx context.Context, resourceType entity.ResourceType) ([]*entity.Resource, error) {
	var models []ResourceModel
	result := r.db.WithContext(ctx).Where("type = ?", string(resourceType)).
		Order("created_at DESC").Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("resourceRepository.ListByType: failed to list resources: %w", result.Error)
	}

	resources := make([]*entity.Resource, len(models))
	for i, model := range models {
		resource, err := toEntity(&model)
		if err != nil {
			return nil, err
		}
		resources[i] = resource
	}
	return resources, nil
}

// UpdateStatus 更新资源状态
func (r *ResourceRepositoryImpl) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ResourceStatus) error {
	result := r.db.WithContext(ctx).Model(&ResourceModel{}).
		Where("id = ?", id.String()).
		Updates(map[string]interface{}{
			"status":     string(status),
			"updated_at": time.Now().Unix(),
		})
	if result.Error != nil {
		return fmt.Errorf("resourceRepository.UpdateStatus: failed to update status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domainErr.ErrNotFound
	}
	return nil
}
