package tool_repository

import (
	"time"

	"gorm.io/datatypes"

	"github.com/google/uuid"

	"mengri-flow/internal/domain/repository"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/register"
)

// ToolModel GORM模型
type ToolModel struct {
	ID          uuid.UUID      `gorm:"type:char(36);primary_key"`
	Name        string         `gorm:"type:varchar(255);not null"`
	Description string         `gorm:"type:text"`
	Type        string         `gorm:"type:varchar(100);not null"`
	Config      datatypes.JSON `gorm:"type:json;not null"`
	ResourceID  uuid.UUID      `gorm:"type:char(36);not null;index:idx_resource"`
	Version     int            `gorm:"type:int;not null;default:1"`
	Status      string         `gorm:"type:varchar(50);not null;default:'draft';index:idx_status"`
	WorkspaceID uuid.UUID      `gorm:"type:char(36);not null;index:idx_workspace"`
	CreatedBy   string         `gorm:"type:varchar(255);not null"`
	CreatedAt   time.Time      `gorm:"type:datetime;not null"`
	UpdatedAt   time.Time      `gorm:"type:datetime;not null"`
}

// TableName 返回表名
func (ToolModel) TableName() string {
	return "tools"
}

// Auto 注册 ToolRepository 工厂及 GORM 自动迁移模型。
func Auto(eventName string) {

	autowire.Auto(func() repository.ToolRepository {

		register.Register(eventName, &ToolModel{})

		return &ToolRepositoryImpl{}
	})
}
