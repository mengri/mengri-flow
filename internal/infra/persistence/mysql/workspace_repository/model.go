package workspaceRepository

import (
	"time"

	"mengri-flow/internal/domain/repository"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/register"
)

func Auto(eventName string) {
	// 注册 WorkspaceRepository 工厂及 GORM 自动迁移模型
	autowire.Auto(func() repository.WorkspaceRepository {
		register.Register(eventName, &WorkspaceModel{})
		return &WorkspaceRepositoryImpl{}
	})
}

// WorkspaceModel GORM 数据库模型
type WorkspaceModel struct {
	ID          string    `gorm:"type:varchar(36);primaryKey"`
	Name        string    `gorm:"type:varchar(100);not null;index:idx_name"`
	Description string    `gorm:"type:text"`
	OwnerID     string    `gorm:"type:varchar(36);not null;index:idx_owner_id"`
	MemberCount int       `gorm:"not null;default:1"`
	CreatedAt   time.Time `gorm:"type:datetime(3);not null;index:idx_created_at"`
	UpdatedAt   time.Time `gorm:"type:datetime(3);not null"`
}

// TableName 指定表名
func (WorkspaceModel) TableName() string {
	return "workspaces"
}
