package auditEventRepository

import (
	"mengri-flow/internal/domain/repository"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/register"
	"time"
)

// Auto 注册 AuditEventRepository 工厂及 GORM 自动迁移模型。
func Auto(eventName string) {
	autowire.Auto(func() repository.AuditEventRepository {
		register.Register(eventName, &AuditEventModel{})
		return &AuditEventRepositoryImpl{}
	})
}

// AuditEventModel GORM 数据库模型。
type AuditEventModel struct {
	ID              string    `gorm:"type:varchar(36);primaryKey"`
	ActorID         string    `gorm:"type:varchar(36);not null;default:'';index:idx_audit_actor"`
	TargetAccountID string    `gorm:"type:varchar(36);not null;default:'';index:idx_audit_target"`
	EventType       string    `gorm:"type:varchar(50);not null;index:idx_audit_event_type"`
	Result          string    `gorm:"type:varchar(10);not null;default:'success'"`
	IP              string    `gorm:"type:varchar(45);not null;default:''"`
	UA              string    `gorm:"type:varchar(500);not null;default:''"`
	MetadataJSON    *string   `gorm:"type:text;column:metadata_json"`
	CreatedAt       time.Time `gorm:"type:datetime(3);not null;index:idx_audit_created_at"`
}

// TableName 指定表名。
func (AuditEventModel) TableName() string {
	return "audit_events"
}
