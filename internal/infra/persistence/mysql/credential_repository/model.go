package credentialRepository

import (
	"mengri-flow/internal/domain/repository"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/register"
	"time"
)

// Auto 注册 CredentialRepository 工厂及 GORM 自动迁移模型。
func Auto(eventName string) {
	autowire.Auto(func() repository.CredentialRepository {
		register.Register(eventName, &CredentialModel{})
		return &CredentialRepositoryImpl{}
	})
}

// CredentialModel GORM 数据库模型。
type CredentialModel struct {
	AccountID         string    `gorm:"type:varchar(36);primaryKey"`
	PasswordHash      string    `gorm:"type:varchar(255);not null"`
	PasswordUpdatedAt time.Time `gorm:"type:datetime(3);not null"`
}

// TableName 指定表名。
func (CredentialModel) TableName() string {
	return "account_credentials"
}
