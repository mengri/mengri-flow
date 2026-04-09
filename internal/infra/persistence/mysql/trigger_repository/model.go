package triggerRepository

import (
	"time"

	"mengri-flow/internal/domain/repository"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/register"
)

func Auto(eventName string) {
	// 注册 TriggerRepository 工厂及 GORM 自动迁移模型
	autowire.Auto(func() repository.TriggerRepository {
		register.Register(eventName, &TriggerModel{})
		return &TriggerRepositoryImpl{}
	})
}

// TriggerModel GORM 数据库模型
type TriggerModel struct {
	ID             string     `gorm:"type:varchar(36);primaryKey"`
	Name           string     `gorm:"type:varchar(100);not null;index:idx_name"`
	Type           string     `gorm:"type:varchar(20);not null;index:idx_type"`
	FlowID         string     `gorm:"type:varchar(36);not null;index:idx_flow_id"`
	FlowVersion    int        `gorm:"type:int;default:1"`
	ClusterID      string     `gorm:"type:varchar(36);index:idx_cluster_id"`
	Config         string     `gorm:"type:json"`
	InputMapping   string     `gorm:"type:json"`
	OutputMapping  string     `gorm:"type:json"`
	ErrorHandling  string     `gorm:"type:json"`
	WorkspaceID    string     `gorm:"type:varchar(36);index:idx_workspace_id"`
	Status         string     `gorm:"type:varchar(20);not null;default:'active';index:idx_status"`
	LastExecutedAt *time.Time `gorm:"type:datetime(3);index:idx_last_executed_at"`
	CreatedAt      time.Time  `gorm:"type:datetime(3);not null;index:idx_created_at"`
	UpdatedAt      time.Time  `gorm:"type:datetime(3);not null"`
}

// TableName 指定表名
func (TriggerModel) TableName() string {
	return "triggers"
}
