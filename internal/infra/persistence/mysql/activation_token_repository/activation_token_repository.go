package activationTokenRepository

import (
	"context"
	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/repository"
	"time"

	"gorm.io/gorm"
)

// ActivationTokenRepositoryImpl 是 ActivationTokenRepository 的 GORM 实现。
type ActivationTokenRepositoryImpl struct {
	db *gorm.DB `autowired:""`
}

var _ repository.ActivationTokenRepository = (*ActivationTokenRepositoryImpl)(nil)

func (r *ActivationTokenRepositoryImpl) Create(ctx context.Context, token *entity.ActivationToken) error {
	model := toModel(token)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *ActivationTokenRepositoryImpl) GetByHash(ctx context.Context, tokenHash string) (*entity.ActivationToken, error) {
	var model ActivationTokenModel
	result := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErr.ErrActivationTokenInvalid
		}
		return nil, result.Error
	}
	return toEntity(&model), nil
}

func (r *ActivationTokenRepositoryImpl) InvalidateByAccountID(ctx context.Context, accountID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&ActivationTokenModel{}).
		Where("account_id = ? AND used_at IS NULL", accountID).
		Update("used_at", now).Error
}

func (r *ActivationTokenRepositoryImpl) MarkUsed(ctx context.Context, tokenHash string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&ActivationTokenModel{}).
		Where("token_hash = ? AND used_at IS NULL", tokenHash).
		Update("used_at", now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainErr.ErrActivationTokenUsed
	}
	return nil
}

func toModel(token *entity.ActivationToken) *ActivationTokenModel {
	return &ActivationTokenModel{
		TokenHash: token.TokenHash,
		AccountID: token.AccountID,
		ExpiresAt: token.ExpiresAt,
		UsedAt:    token.UsedAt,
		CreatedAt: token.CreatedAt,
	}
}

func toEntity(model *ActivationTokenModel) *entity.ActivationToken {
	return &entity.ActivationToken{
		TokenHash: model.TokenHash,
		AccountID: model.AccountID,
		ExpiresAt: model.ExpiresAt,
		UsedAt:    model.UsedAt,
		CreatedAt: model.CreatedAt,
	}
}
