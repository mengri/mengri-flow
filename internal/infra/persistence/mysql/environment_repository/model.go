package environmentRepository

import (
	"time"

	"mengri-flow/internal/domain/repository"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/register"
)

// Auto 注册 EnvironmentRepository 工厂及 GORM 自动迁移模型
func Auto(eventName string) {
	autowire.Auto(func() repository.EnvironmentRepository {
		register.Register(eventName, &EnvironmentModel{})
		return &EnvironmentRepositoryImpl{}
	})
}

// EnvironmentModel GORM 数据库模型
type EnvironmentModel struct {
	ID          string    `gorm:"type:varchar(36);primaryKey"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Key         string    `gorm:"type:varchar(50);uniqueIndex:uk_key;not null"`
	Description string    `gorm:"type:text"`
	Color       string    `gorm:"type:varchar(7)"` // Hex颜色，如 #FF5733
	CreatedAt   time.Time `gorm:"type:datetime(3);not null;index:idx_created_at"`
	UpdatedAt   time.Time `gorm:"type:datetime(3);not null"`
}

// TableName 指定表名
func (EnvironmentModel) TableName() string {
	return "environments"
}
