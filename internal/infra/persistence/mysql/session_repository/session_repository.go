package sessionRepository

import (
	"context"
	"mengri-flow/internal/domain/repository"
	"time"

	domainErr "mengri-flow/internal/domain/errors"

	"gorm.io/gorm"
)

// SessionStoreImpl 是 SessionStore 的 GORM 实现。
type SessionStoreImpl struct {
	db *gorm.DB `autowired:""`
}

var _ repository.SessionStore = (*SessionStoreImpl)(nil)

func (r *SessionStoreImpl) SaveRefreshToken(ctx context.Context, sessionID, accountID, tokenHash, deviceInfoJSON, ip string, ttl time.Duration) error {
	var deviceInfo *string
	if deviceInfoJSON != "" {
		deviceInfo = &deviceInfoJSON
	}
	model := &SessionModel{
		ID:               sessionID,
		AccountID:        accountID,
		RefreshTokenHash: tokenHash,
		DeviceInfoJSON:   deviceInfo,
		IP:               ip,
		ExpiresAt:        time.Now().Add(ttl),
		CreatedAt:        time.Now(),
	}
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *SessionStoreImpl) ValidateRefreshToken(ctx context.Context, tokenHash string) (string, error) {
	var model SessionModel
	result := r.db.WithContext(ctx).
		Where("refresh_token_hash = ? AND expires_at > ?", tokenHash, time.Now()).
		First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return "", domainErr.ErrSessionExpired
		}
		return "", result.Error
	}
	return model.AccountID, nil
}

func (r *SessionStoreImpl) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	return r.db.WithContext(ctx).
		Where("refresh_token_hash = ?", tokenHash).
		Delete(&SessionModel{}).Error
}

func (r *SessionStoreImpl) RevokeAllByAccountID(ctx context.Context, accountID string, exceptTokenHash string) (int, error) {
	result := r.db.WithContext(ctx).
		Where("account_id = ? AND refresh_token_hash != ?", accountID, exceptTokenHash).
		Delete(&SessionModel{})
	return int(result.RowsAffected), result.Error
}
