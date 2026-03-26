package auditEventRepository

import (
	"context"
	"mengri-flow/internal/domain/entity"
	"mengri-flow/internal/domain/repository"

	"gorm.io/gorm"
)

// AuditEventRepositoryImpl 是 AuditEventRepository 的 GORM 实现。
type AuditEventRepositoryImpl struct {
	db *gorm.DB `autowired:""`
}

var _ repository.AuditEventRepository = (*AuditEventRepositoryImpl)(nil)

func (r *AuditEventRepositoryImpl) Create(ctx context.Context, event *entity.AuditEvent) error {
	model := toModel(event)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *AuditEventRepositoryImpl) ListByAccountID(ctx context.Context, accountID string, offset, limit int) ([]*entity.AuditEvent, int64, error) {
	var models []AuditEventModel
	var total int64

	query := r.db.WithContext(ctx).Model(&AuditEventModel{}).Where("target_account_id = ?", accountID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	events := make([]*entity.AuditEvent, 0, len(models))
	for _, m := range models {
		events = append(events, toEntity(&m))
	}
	return events, total, nil
}

func (r *AuditEventRepositoryImpl) ListByFilter(ctx context.Context, filter repository.AuditFilter) ([]*entity.AuditEvent, int64, error) {
	var models []AuditEventModel
	var total int64

	query := r.db.WithContext(ctx).Model(&AuditEventModel{})
	if filter.AccountID != "" {
		query = query.Where("target_account_id = ?", filter.AccountID)
	}
	if filter.EventType != "" {
		query = query.Where("event_type = ?", filter.EventType)
	}
	if filter.From != nil {
		query = query.Where("created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		query = query.Where("created_at <= ?", *filter.To)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(filter.Offset).Limit(filter.Limit).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	events := make([]*entity.AuditEvent, 0, len(models))
	for _, m := range models {
		events = append(events, toEntity(&m))
	}
	return events, total, nil
}

func toModel(event *entity.AuditEvent) *AuditEventModel {
	var meta *string
	if event.Metadata != "" {
		meta = &event.Metadata
	}
	return &AuditEventModel{
		ID:              event.ID,
		ActorID:         event.ActorID,
		TargetAccountID: event.TargetAccountID,
		EventType:       event.EventType,
		Result:          event.Result,
		IP:              event.IP,
		UA:              event.UA,
		MetadataJSON:    meta,
		CreatedAt:       event.CreatedAt,
	}
}

func toEntity(model *AuditEventModel) *entity.AuditEvent {
	meta := ""
	if model.MetadataJSON != nil {
		meta = *model.MetadataJSON
	}
	return &entity.AuditEvent{
		ID:              model.ID,
		ActorID:         model.ActorID,
		TargetAccountID: model.TargetAccountID,
		EventType:       model.EventType,
		Result:          model.Result,
		IP:              model.IP,
		UA:              model.UA,
		Metadata:        meta,
		CreatedAt:       model.CreatedAt,
	}
}
