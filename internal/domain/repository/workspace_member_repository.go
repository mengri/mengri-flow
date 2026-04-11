package repository

import (
	"context"

	"github.com/google/uuid"

	"mengri-flow/internal/domain/entity"
)

// WorkspaceMemberRepository 工作空间成员仓储接口
type WorkspaceMemberRepository interface {
	// Create 创建成员关系
	Create(ctx context.Context, member *entity.WorkspaceMember) error
	// Delete 删除成员关系
	Delete(ctx context.Context, workspaceID uuid.UUID, accountID string) error
	// FindByWorkspaceIDAndAccountID 查询指定工作空间和账号的成员关系
	FindByWorkspaceIDAndAccountID(ctx context.Context, workspaceID uuid.UUID, accountID string) (*entity.WorkspaceMember, error)
	// ListByWorkspaceID 分页查询工作空间的成员列表
	ListByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, offset, limit int) ([]*entity.WorkspaceMember, int64, error)
	// CountByWorkspaceID 查询工作空间的成员总数
	CountByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error)
}
