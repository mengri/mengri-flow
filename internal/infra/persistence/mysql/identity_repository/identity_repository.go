package identityRepository

import (
	"context"
	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/repository"
	"time"

	"gorm.io/gorm"
)

// IdentityRepositoryImpl 是 IdentityRepository 的 GORM 实现。
type IdentityRepositoryImpl struct {
	db *gorm.DB `autowired:""`
}

var _ repository.IdentityRepository = (*IdentityRepositoryImpl)(nil)

func (r *IdentityRepositoryImpl) Create(ctx context.Context, identity *entity.Identity) error {
	model := toModel(identity)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *IdentityRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.Identity, error) {
	var model IdentityModel
	result := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErr.ErrIdentityNotBound
		}
		return nil, result.Error
	}
	return toEntity(&model), nil
}

func (r *IdentityRepositoryImpl) GetByProviderID(ctx context.Context, loginType entity.LoginType, externalID string) (*entity.Identity, error) {
	var model IdentityModel
	result := r.db.WithContext(ctx).
		Where("login_type = ? AND external_id = ? AND deleted_at IS NULL", string(loginType), externalID).
		First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErr.ErrIdentityNotBound
		}
		return nil, result.Error
	}
	return toEntity(&model), nil
}

func (r *IdentityRepositoryImpl) ListByAccountID(ctx context.Context, accountID string) ([]*entity.Identity, error) {
	var models []IdentityModel
	err := r.db.WithContext(ctx).
		Where("account_id = ? AND deleted_at IS NULL", accountID).
		Order("created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	identities := make([]*entity.Identity, 0, len(models))
	for _, m := range models {
		identities = append(identities, toEntity(&m))
	}
	return identities, nil
}

func (r *IdentityRepositoryImpl) CountActiveByAccountID(ctx context.Context, accountID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&IdentityModel{}).
		Where("account_id = ? AND deleted_at IS NULL", accountID).
		Count(&count).Error
	return int(count), err
}

func (r *IdentityRepositoryImpl) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&IdentityModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainErr.ErrIdentityNotBound
	}
	return nil
}

func toModel(identity *entity.Identity) *IdentityModel {
	var meta *string
	if identity.ExternalMeta != "" {
		meta = &identity.ExternalMeta
	}
	return &IdentityModel{
		ID:               identity.ID,
		AccountID:        identity.AccountID,
		LoginType:        string(identity.LoginType),
		ExternalID:       identity.ExternalID,
		ExternalMetaJSON: meta,
		CreatedAt:        identity.CreatedAt,
		DeletedAt:        identity.DeletedAt,
	}
}

func toEntity(model *IdentityModel) *entity.Identity {
	meta := ""
	if model.ExternalMetaJSON != nil {
		meta = *model.ExternalMetaJSON
	}
	return &entity.Identity{
		ID:           model.ID,
		AccountID:    model.AccountID,
		LoginType:    entity.LoginType(model.LoginType),
		ExternalID:   model.ExternalID,
		ExternalMeta: meta,
		CreatedAt:    model.CreatedAt,
		DeletedAt:    model.DeletedAt,
	}
}
