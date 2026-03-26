package activationTokenRepository

import (
	"mengri-flow/internal/domain/repository"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/register"
	"time"
)

// Auto 注册 ActivationTokenRepository 工厂及 GORM 自动迁移模型。
func Auto(eventName string) {
	autowire.Auto(func() repository.ActivationTokenRepository {
		register.Register(eventName, &ActivationTokenModel{})
		return &ActivationTokenRepositoryImpl{}
	})
}

// ActivationTokenModel GORM 数据库模型。
type ActivationTokenModel struct {
	TokenHash string     `gorm:"type:varchar(64);primaryKey"`
	AccountID string     `gorm:"type:varchar(36);not null;index:idx_act_account_id"`
	ExpiresAt time.Time  `gorm:"type:datetime(3);not null;index:idx_act_expires_at"`
	UsedAt    *time.Time `gorm:"type:datetime(3)"`
	CreatedAt time.Time  `gorm:"type:datetime(3);not null"`
}

// TableName 指定表名。
func (ActivationTokenModel) TableName() string {
	return "activation_tokens"
}
