package triggerRepository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/repository"

	"gorm.io/gorm"
)

// TriggerRepositoryImpl 是 TriggerRepository 的 GORM实现
type TriggerRepositoryImpl struct {
	db *gorm.DB `autowired:""`
}

var _ repository.TriggerRepository = (*TriggerRepositoryImpl)(nil)

// Create 创建触发器
func (r *TriggerRepositoryImpl) Create(ctx context.Context, trigger *entity.Trigger) error {
	model := toModel(trigger)
	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return fmt.Errorf("triggerRepository.Create: failed to create trigger: %w", result.Error)
	}
	trigger.ID = uuid.MustParse(model.ID)
	return nil
}

// Update 更新触发器
func (r *TriggerRepositoryImpl) Update(ctx context.Context, trigger *entity.Trigger) error {
	model := toModel(trigger)
	result := r.db.WithContext(ctx).Model(&TriggerModel{}).Where("id = ?", model.ID).Updates(model)
	if result.Error != nil {
		return fmt.Errorf("triggerRepository.Update: failed to update trigger: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domainErr.ErrNotFound
	}
	return nil
}

// Delete 删除触发器
func (r *TriggerRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&TriggerModel{}, "id = ?", id.String())
	if result.Error != nil {
		return fmt.Errorf("triggerRepository.Delete: failed to delete trigger: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domainErr.ErrNotFound
	}
	return nil
}

// FindByID 根据ID查找触发器
func (r *TriggerRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.Trigger, error) {
	var model TriggerModel
	result := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErr.ErrNotFound
		}
		return nil, fmt.Errorf("triggerRepository.FindByID: failed to find trigger: %w", result.Error)
	}
	return toEntity(&model)
}

// FindByFlowID 根据FlowID查找触发器列表
func (r *TriggerRepositoryImpl) FindByFlowID(ctx context.Context, flowID uuid.UUID) ([]*entity.Trigger, error) {
	var models []TriggerModel
	result := r.db.WithContext(ctx).Where("flow_id = ?", flowID.String()).
		Order("created_at DESC").Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("triggerRepository.FindByFlowID: failed to find triggers: %w", result.Error)
	}

	triggers := make([]*entity.Trigger, len(models))
	for i, model := range models {
		trigger, err := toEntity(&model)
		if err != nil {
			return nil, err
		}
		triggers[i] = trigger
	}
	return triggers, nil
}

// ListWithFilters 根据条件列出触发器
func (r *TriggerRepositoryImpl) ListWithFilters(ctx context.Context, flowID *string, status *entity.TriggerStatus, offset, limit int) ([]*entity.Trigger, int64, error) {
	var models []TriggerModel
	var total int64

	query := r.db.WithContext(ctx).Model(&TriggerModel{})

	if flowID != nil {
		query = query.Where("flow_id = ?", *flowID)
	}
	if status != nil {
		query = query.Where("status = ?", string(*status))
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("triggerRepository.ListWithFilters: failed to count triggers: %w", err)
	}

	// 获取分页数据
	result := query.Order("created_at DESC").
		Offset(offset).Limit(limit).Find(&models)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("triggerRepository.ListWithFilters: failed to list triggers: %w", result.Error)
	}

	triggers := make([]*entity.Trigger, len(models))
	for i, model := range models {
		trigger, err := toEntity(&model)
		if err != nil {
			return nil, 0, err
		}
		triggers[i] = trigger
	}
	return triggers, total, nil
}