package workspaceMemberRepository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/repository"
)

// WorkspaceMemberRepositoryImpl 是 WorkspaceMemberRepository 的 GORM 实现
type WorkspaceMemberRepositoryImpl struct {
	db *gorm.DB `autowired:""`
}

var _ repository.WorkspaceMemberRepository = (*WorkspaceMemberRepositoryImpl)(nil)

// Create 创建成员关系
func (r *WorkspaceMemberRepositoryImpl) Create(ctx context.Context, member *entity.WorkspaceMember) error {
	model := toModel(member)
	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		if isDuplicateKeyError(result.Error) {
			return domainErr.ErrConflict
		}
		return fmt.Errorf("workspaceMemberRepository.Create: %w", result.Error)
	}
	return nil
}

// Delete 删除成员关系
func (r *WorkspaceMemberRepositoryImpl) Delete(ctx context.Context, workspaceID uuid.UUID, accountID string) error {
	result := r.db.WithContext(ctx).
		Where("workspace_id = ? AND account_id = ?", workspaceID.String(), accountID).
		Delete(&WorkspaceMemberModel{})
	if result.Error != nil {
		return fmt.Errorf("workspaceMemberRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domainErr.ErrNotFound
	}
	return nil
}

// FindByWorkspaceIDAndAccountID 查询指定工作空间和账号的成员关系
func (r *WorkspaceMemberRepositoryImpl) FindByWorkspaceIDAndAccountID(ctx context.Context, workspaceID uuid.UUID, accountID string) (*entity.WorkspaceMember, error) {
	var model WorkspaceMemberModel
	result := r.db.WithContext(ctx).
		Where("workspace_id = ? AND account_id = ?", workspaceID.String(), accountID).
		First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErr.ErrNotFound
		}
		return nil, fmt.Errorf("workspaceMemberRepository.FindByWorkspaceIDAndAccountID: %w", result.Error)
	}
	return toEntity(&model), nil
}

// ListByWorkspaceID 分页查询工作空间的成员列表
func (r *WorkspaceMemberRepositoryImpl) ListByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, offset, limit int) ([]*entity.WorkspaceMember, int64, error) {
	var models []WorkspaceMemberModel
	var total int64

	if err := r.db.WithContext(ctx).
		Model(&WorkspaceMemberModel{}).
		Where("workspace_id = ?", workspaceID.String()).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("workspaceMemberRepository.ListByWorkspaceID: count: %w", err)
	}

	result := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID.String()).
		Order("joined_at ASC").
		Offset(offset).Limit(limit).
		Find(&models)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("workspaceMemberRepository.ListByWorkspaceID: %w", result.Error)
	}

	members := make([]*entity.WorkspaceMember, len(models))
	for i, m := range models {
		members[i] = toEntity(&m)
	}
	return members, total, nil
}

// CountByWorkspaceID 查询工作空间的成员总数
func (r *WorkspaceMemberRepositoryImpl) CountByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&WorkspaceMemberModel{}).
		Where("workspace_id = ?", workspaceID.String()).
		Count(&count).Error
	return count, err
}

// isDuplicateKeyError 判断是否为唯一键冲突错误（MySQL 1062）
func isDuplicateKeyError(err error) bool {
	return err != nil &&
		(gorm.ErrDuplicatedKey == err ||
			containsString(err.Error(), "Duplicate entry") ||
			containsString(err.Error(), "1062"))
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
