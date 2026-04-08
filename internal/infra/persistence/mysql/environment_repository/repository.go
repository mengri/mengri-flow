package environmentRepository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/repository"

	"gorm.io/gorm"
)

// EnvironmentRepositoryImpl 是 EnvironmentRepository 的 GORM 实现
type EnvironmentRepositoryImpl struct {
	db *gorm.DB `autowired:""`
}

var _ repository.EnvironmentRepository = (*EnvironmentRepositoryImpl)(nil)

// Create 创建环境
func (r *EnvironmentRepositoryImpl) Create(ctx context.Context, env *entity.Environment) error {
	model := toModel(env)
	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return fmt.Errorf("environmentRepository.Create: failed to create environment: %w", result.Error)
	}
	env.ID = uuid.MustParse(model.ID)
	return nil
}

// Update 更新环境
func (r *EnvironmentRepositoryImpl) Update(ctx context.Context, env *entity.Environment) error {
	model := toModel(env)
	result := r.db.WithContext(ctx).Model(&EnvironmentModel{}).Where("id = ?", model.ID).Updates(model)
	if result.Error != nil {
		return fmt.Errorf("environmentRepository.Update: failed to update environment: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domainErr.ErrNotFound
	}
	return nil
}

// Delete 删除环境
func (r *EnvironmentRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&EnvironmentModel{}, "id = ?", id.String())
	if result.Error != nil {
		return fmt.Errorf("environmentRepository.Delete: failed to delete environment: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domainErr.ErrNotFound
	}
	return nil
}

// FindByID 根据ID查找环境
func (r *EnvironmentRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.Environment, error) {
	var model EnvironmentModel
	result := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErr.ErrNotFound
		}
		return nil, fmt.Errorf("environmentRepository.FindByID: failed to find environment: %w", result.Error)
	}
	return toEntity(&model), nil
}

// FindByKey 根据Key查找环境
func (r *EnvironmentRepositoryImpl) FindByKey(ctx context.Context, key string) (*entity.Environment, error) {
	var model EnvironmentModel
	result := r.db.WithContext(ctx).Where("key = ?", key).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErr.ErrNotFound
		}
		return nil, fmt.Errorf("environmentRepository.FindByKey: failed to find environment: %w", result.Error)
	}
	return toEntity(&model), nil
}

// List 列出所有环境
func (r *EnvironmentRepositoryImpl) List(ctx context.Context) ([]*entity.Environment, error) {
	var models []EnvironmentModel
	result := r.db.WithContext(ctx).Order("created_at DESC").Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("environmentRepository.List: failed to list environments: %w", result.Error)
	}

	environments := make([]*entity.Environment, len(models))
	for i, model := range models {
		environments[i] = toEntity(&model)
	}
	return environments, nil
}
