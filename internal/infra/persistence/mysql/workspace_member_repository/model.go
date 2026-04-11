package workspaceMemberRepository

import (
	"mengri-flow/internal/domain/repository"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/register"
	"time"
)

// Auto 注册 WorkspaceMemberRepository 工厂及 GORM 自动迁移模型
func Auto(eventName string) {
	autowire.Auto(func() repository.WorkspaceMemberRepository {
		register.Register(eventName, &WorkspaceMemberModel{})
		return &WorkspaceMemberRepositoryImpl{}
	})
}

// WorkspaceMemberModel GORM 数据库模型
type WorkspaceMemberModel struct {
	WorkspaceID string    `gorm:"type:varchar(36);not null;uniqueIndex:uk_workspace_account;primaryKey"`
	AccountID   string    `gorm:"type:varchar(36);not null;uniqueIndex:uk_workspace_account"`
	Role        string    `gorm:"type:varchar(20);not null;default:member"`
	JoinedAt    time.Time `gorm:"type:datetime(3);not null"`
}

// TableName 指定表名
func (WorkspaceMemberModel) TableName() string {
	return "workspace_members"
}
