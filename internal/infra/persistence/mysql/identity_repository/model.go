package identityRepository

import (
	"mengri-flow/internal/domain/repository"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/register"
	"time"
)

// Auto 注册 IdentityRepository 工厂及 GORM 自动迁移模型。
func Auto(eventName string) {
	autowire.Auto(func() repository.IdentityRepository {
		register.Register(eventName, &IdentityModel{})
		return &IdentityRepositoryImpl{}
	})
}

// IdentityModel GORM 数据库模型。
type IdentityModel struct {
	ID               string     `gorm:"type:varchar(36);primaryKey"`
	AccountID        string     `gorm:"type:varchar(36);not null;index:idx_account_id;uniqueIndex:uk_account_login_type"`
	LoginType        string     `gorm:"type:varchar(30);not null;uniqueIndex:uk_login_type_external_id;uniqueIndex:uk_account_login_type"`
	ExternalID       string     `gorm:"type:varchar(255);not null;uniqueIndex:uk_login_type_external_id"`
	ExternalMetaJSON *string    `gorm:"type:text;column:external_meta_json"`
	CreatedAt        time.Time  `gorm:"type:datetime(3);not null"`
	DeletedAt        *time.Time `gorm:"type:datetime(3);index"`
}

// TableName 指定表名。
func (IdentityModel) TableName() string {
	return "account_identities"
}
