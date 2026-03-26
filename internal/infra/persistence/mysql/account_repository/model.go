package accountRepository

import (
	"mengri-flow/internal/domain/repository"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/register"
	"time"
)

// Auto 注册 AccountRepository 工厂及 GORM 自动迁移模型。
func Auto(eventName string) {
	autowire.Auto(func() repository.AccountRepository {
		register.Register(eventName, &AccountModel{})
		return &AccountRepositoryImpl{}
	})
}

// AccountModel GORM 数据库模型。
type AccountModel struct {
	ID          string     `gorm:"type:varchar(36);primaryKey"`
	Email       string     `gorm:"type:varchar(255);uniqueIndex:uk_email;not null"`
	Username    string     `gorm:"type:varchar(50);uniqueIndex:uk_username;not null"`
	DisplayName string     `gorm:"type:varchar(100);not null;default:''"`
	Status      string     `gorm:"type:varchar(30);not null;default:'PENDING_ACTIVATION';index:idx_status"`
	Role        string     `gorm:"type:varchar(20);not null;default:'user'"`
	ActivatedAt *time.Time `gorm:"type:datetime(3)"`
	CreatedAt   time.Time  `gorm:"type:datetime(3);not null;index:idx_created_at"`
	UpdatedAt   time.Time  `gorm:"type:datetime(3);not null"`
}

// TableName 指定表名。
func (AccountModel) TableName() string {
	return "accounts"
}
