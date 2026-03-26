package sessionRepository

import (
	"mengri-flow/internal/domain/repository"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/register"
	"time"
)

// Auto 注册 SessionStore 工厂及 GORM 自动迁移模型。
func Auto(eventName string) {
	autowire.Auto(func() repository.SessionStore {
		register.Register(eventName, &SessionModel{})
		return &SessionStoreImpl{}
	})
}

// SessionModel GORM 数据库模型。
type SessionModel struct {
	ID               string    `gorm:"type:varchar(36);primaryKey"`
	AccountID        string    `gorm:"type:varchar(36);not null;index:idx_sess_account_id"`
	RefreshTokenHash string    `gorm:"type:varchar(64);not null;index:idx_sess_refresh_token"`
	DeviceInfoJSON   *string   `gorm:"type:text;column:device_info_json"`
	IP               string    `gorm:"type:varchar(45);not null;default:''"`
	ExpiresAt        time.Time `gorm:"type:datetime(3);not null;index:idx_sess_expires_at"`
	CreatedAt        time.Time `gorm:"type:datetime(3);not null"`
}

// TableName 指定表名。
func (SessionModel) TableName() string {
	return "sessions"
}
