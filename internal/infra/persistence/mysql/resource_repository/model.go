package resourceRepository

import (
	"mengri-flow/internal/domain/repository"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/register"
)

// Auto 注册 ResourceRepository 工厂及 GORM 自动迁移模型
func Auto(eventName string) {
	autowire.Auto(func() repository.ResourceRepository {
		register.Register(eventName, &ResourceModel{})
		return &ResourceRepositoryImpl{}
	})
}

// ResourceModel GORM 数据库模型
type ResourceModel struct {
	ID          string `gorm:"type:varchar(36);primaryKey"`
	Name        string `gorm:"type:varchar(100);not null;index:idx_name"`
	Type        string `gorm:"type:varchar(20);not null;index:idx_type"`
	Config      string `gorm:"type:json"`
	WorkspaceID string `gorm:"type:varchar(36);not null;index:idx_workspace_id"`
	Status      string `gorm:"type:varchar(20);not null;default:'active';index:idx_status"`
	Description string `gorm:"type:text"`
	CreatedAt   int64  `gorm:"not null;index:idx_created_at"`
	UpdatedAt   int64  `gorm:"not null"`
}

// TableName 指定表名
func (ResourceModel) TableName() string {
	return "resources"
}
