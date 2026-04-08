package clusterRepository

import (
	"time"

	"mengri-flow/internal/domain/repository"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/register"
)

// Auto 注册 ClusterRepository 工厂及 GORM 自动迁移模型
func Auto(eventName string) {
	autowire.Auto(func() repository.ClusterRepository {
		register.Register(eventName, &ClusterModel{})
		return &ClusterRepositoryImpl{}
	})
}

// ClusterModel GORM 数据库模型
type ClusterModel struct {
	ID             string     `gorm:"type:varchar(36);primaryKey"`
	Name           string     `gorm:"type:varchar(100);not null;index:idx_name"`
	EnvironmentID  string     `gorm:"type:varchar(36);not null;index:idx_environment_id"`
	EtcdEndpoints  string     `gorm:"type:text"` // JSON serialized
	EtcdUsername   string     `gorm:"type:varchar(255)"`
	EtcdPassword   string     `gorm:"type:varchar(255)"` // 加密存储
	EtcdPrefix     string     `gorm:"type:varchar(255)"`
	Description    string     `gorm:"type:text"`
	Status         string     `gorm:"type:varchar(20);not null;default:'active';index:idx_status"`
	ExecutorCount  int        `gorm:"not null;default:0"`
	LastHeartbeat  *time.Time `gorm:"type:datetime(3);index:idx_last_heartbeat"`
	CreatedAt      time.Time  `gorm:"type:datetime(3);not null;index:idx_created_at"`
	UpdatedAt      time.Time  `gorm:"type:datetime(3);not null"`
}

// TableName 指定表名
func (ClusterModel) TableName() string {
	return "clusters"
}
