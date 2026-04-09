package flow_repository

import (
	"time"

	"gorm.io/datatypes"

	"github.com/google/uuid"

	"mengri-flow/internal/domain/repository"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/register"
)

// FlowModel GORM模型
type FlowModel struct {
	ID          uuid.UUID      `gorm:"type:char(36);primary_key"`
	Name        string         `gorm:"type:varchar(255);not null"`
	Description string         `gorm:"type:text"`
	CanvasData  datatypes.JSON `gorm:"type:json;not null"`
	Status      string         `gorm:"type:varchar(50);not null;default:'draft';index:idx_status"`
	Version     int            `gorm:"type:int;not null;default:1"`
	WorkspaceID uuid.UUID      `gorm:"type:char(36);not null;index:idx_workspace"`
	CreatedBy   string         `gorm:"type:varchar(255);not null"`
	CreatedAt   time.Time      `gorm:"type:datetime;not null"`
	UpdatedAt   time.Time      `gorm:"type:datetime;not null"`
}

// TableName 返回表名
func (FlowModel) TableName() string {
	return "flows"
}

// FlowVersionModel 流程版本模型（用于版本管理）
type FlowVersionModel struct {
	ID          uuid.UUID      `gorm:"type:char(36);primary_key"`
	FlowID      uuid.UUID      `gorm:"type:char(36);not null;index:idx_flow"`
	Name        string         `gorm:"type:varchar(255);not null"`
	Description string         `gorm:"type:text"`
	CanvasData  datatypes.JSON `gorm:"type:json;not null"`
	Status      string         `gorm:"type:varchar(50);not null"`
	Version     int            `gorm:"type:int;not null"`
	WorkspaceID uuid.UUID      `gorm:"type:char(36);not null"`
	CreatedBy   string         `gorm:"type:varchar(255);not null"`
	CreatedAt   time.Time      `gorm:"type:datetime;not null"`
	UpdatedAt   time.Time      `gorm:"type:datetime;not null"`
}

// TableName 返回表名
func (FlowVersionModel) TableName() string {
	return "flow_versions"
}
func Auto(eventName string) {
	autowire.Auto(func() repository.FlowRepository {
		register.Register(eventName, &FlowModel{})
		register.Register(eventName, &FlowVersionModel{})
		return &FlowRepositoryImpl{}
	})
}
