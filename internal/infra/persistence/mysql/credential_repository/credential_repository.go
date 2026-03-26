package credentialRepository

import (
	"context"
	"mengri-flow/internal/domain/repository"
	"time"

	domainErr "mengri-flow/internal/domain/errors"

	"gorm.io/gorm"
)

// CredentialRepositoryImpl 是 CredentialRepository 的 GORM 实现。
type CredentialRepositoryImpl struct {
	db *gorm.DB `autowired:""`
}

var _ repository.CredentialRepository = (*CredentialRepositoryImpl)(nil)

func (r *CredentialRepositoryImpl) Create(ctx context.Context, accountID string, passwordHash string) error {
	model := &CredentialModel{
		AccountID:         accountID,
		PasswordHash:      passwordHash,
		PasswordUpdatedAt: time.Now(),
	}
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *CredentialRepositoryImpl) GetByAccountID(ctx context.Context, accountID string) (string, error) {
	var model CredentialModel
	result := r.db.WithContext(ctx).Where("account_id = ?", accountID).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return "", domainErr.ErrAccountNotFound
		}
		return "", result.Error
	}
	return model.PasswordHash, nil
}

func (r *CredentialRepositoryImpl) UpdatePassword(ctx context.Context, accountID string, passwordHash string) error {
	return r.db.WithContext(ctx).
		Model(&CredentialModel{}).
		Where("account_id = ?", accountID).
		Updates(map[string]interface{}{
			"password_hash":       passwordHash,
			"password_updated_at": time.Now(),
		}).Error
}
