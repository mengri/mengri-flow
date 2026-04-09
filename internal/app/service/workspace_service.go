package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/repository"
)

type workspaceServiceImpl struct {
	workspaceRepo repository.WorkspaceRepository `autowired:""`
	accountRepo   repository.AccountRepository   `autowired:""`
}

func (s *workspaceServiceImpl) CreateWorkspace(ctx context.Context, req *dto.CreateWorkspaceRequest, ownerID string) (*dto.WorkspaceResponse, error) {
	// 1. 验证所有者存在
	owner, err := s.accountRepo.GetByID(ctx, ownerID)
	if err != nil {
		slog.Error("failed to get owner account", "ownerID", ownerID, "error", err)
		return nil, domainErr.ErrNotFound
	}

	// 2. 创建Workspace领域实体
	workspace, err := entity.NewWorkspace(req.Name, req.Description, owner.ID)
	if err != nil {
		return nil, fmt.Errorf("workspace validation failed: %w", err)
	}

	// 3. 保存到数据库
	if err := s.workspaceRepo.Create(ctx, workspace); err != nil {
		slog.Error("failed to create workspace", "workspace", workspace.Name, "error", err)
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// 4. 转换为DTO返回
	return toWorkspaceResponse(workspace), nil
}

func (s *workspaceServiceImpl) ListWorkspaces(ctx context.Context, accountID string, page int, pageSize int) (*dto.ListWorkspacesResponse, error) {
	slog.Info("Listing workspaces", "accountId", accountID, "page", page, "pageSize", pageSize)

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// 使用仓库的分页查询
	workspaces, total, err := s.workspaceRepo.List(ctx, offset, pageSize)
	if err != nil {
		slog.Error("Failed to list workspaces", "error", err)
		return nil, fmt.Errorf("list workspaces: %w", err)
	}

	// 转换为响应DTO
	responseList := make([]dto.WorkspaceResponse, 0, len(workspaces))
	for _, workspace := range workspaces {
		// 检查用户是否有权限查看此工作空间
		// 这里简化处理：只允许用户查看自己拥有的工作空间
		// 或者需要实现工作空间成员查询
		if workspace.OwnerID == accountID {
			responseList = append(responseList, *toWorkspaceResponse(workspace))
		}
	}

	slog.Info("Workspaces listed successfully", "total", total, "returned", len(responseList))
	return &dto.ListWorkspacesResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		List:     responseList,
	}, nil
}

func (s *workspaceServiceImpl) GetWorkspace(ctx context.Context, id string, accountID string) (*dto.WorkspaceResponse, error) {
	slog.Info("Getting workspace", "workspaceId", id, "accountId", accountID)

	// 验证参数
	if id == "" {
		return nil, domainErr.ErrInvalidInput
	}

	workspaceID, err := uuid.Parse(id)
	if err != nil {
		slog.Warn("Invalid workspace ID", "workspaceId", id, "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	// 查找工作空间
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Workspace not found", "workspaceId", workspaceID)
			return nil, domainErr.ErrNotFound
		}
		slog.Error("Failed to find workspace", "workspaceId", workspaceID, "error", err)
		return nil, fmt.Errorf("find workspace: %w", err)
	}

	// 验证权限：只有所有者或成员可以查看
	// 目前只检查是否是所有者，成员检查需要后续实现
	if workspace.OwnerID != accountID {
		// TODO: 实现成员权限检查
		slog.Warn("User does not have permission to view workspace",
			"workspaceId", workspaceID, "userId", accountID, "ownerId", workspace.OwnerID)
		return nil, domainErr.ErrForbidden
	}

	slog.Info("Workspace retrieved successfully", "workspaceId", workspaceID)
	return toWorkspaceResponse(workspace), nil
}

func (s *workspaceServiceImpl) UpdateWorkspace(ctx context.Context, id string, req *dto.UpdateWorkspaceRequest, accountID string) (*dto.WorkspaceResponse, error) {
	// 1. 验证参数
	if id == "" {
		return nil, domainErr.ErrInvalidInput
	}

	workspaceID, err := uuid.Parse(id)
	if err != nil {
		return nil, domainErr.ErrInvalidInput
	}

	// 2. 获取现有工作空间
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			return nil, domainErr.ErrNotFound
		}
		slog.Error("failed to find workspace", "workspaceID", id, "error", err)
		return nil, fmt.Errorf("failed to find workspace: %w", err)
	}

	// 3. 验证权限 - 只有所有者可以更新
	if workspace.OwnerID != accountID {
		return nil, domainErr.ErrForbidden
	}

	// 4. 更新工作空间信息
	if req.Name != "" || req.Description != "" {
		newName := req.Name
		if newName == "" {
			newName = workspace.Name
		}
		newDescription := req.Description
		if newDescription == "" {
			newDescription = workspace.Description
		}

		if err := workspace.Update(newName, newDescription); err != nil {
			return nil, fmt.Errorf("workspace validation failed: %w", err)
		}

		// 5. 保存更新
		if err := s.workspaceRepo.Update(ctx, workspace); err != nil {
			slog.Error("failed to update workspace", "workspaceID", id, "error", err)
			return nil, fmt.Errorf("failed to update workspace: %w", err)
		}
	}

	// 6. 返回更新后的工作空间信息
	return toWorkspaceResponse(workspace), nil
}

func (s *workspaceServiceImpl) DeleteWorkspace(ctx context.Context, id string, accountID string) error {
	slog.Info("Deleting workspace", "workspaceId", id, "accountId", accountID)

	// 验证参数
	if id == "" {
		return domainErr.ErrInvalidInput
	}

	workspaceID, err := uuid.Parse(id)
	if err != nil {
		slog.Warn("Invalid workspace ID", "workspaceId", id, "error", err)
		return domainErr.ErrInvalidInput
	}

	// 查找工作空间
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Workspace not found", "workspaceId", workspaceID)
			return domainErr.ErrNotFound
		}
		slog.Error("Failed to find workspace", "workspaceId", workspaceID, "error", err)
		return fmt.Errorf("find workspace: %w", err)
	}

	// 验证权限：只有所有者可以删除
	if workspace.OwnerID != accountID {
		slog.Warn("User does not have permission to delete workspace",
			"workspaceId", workspaceID, "userId", accountID, "ownerId", workspace.OwnerID)
		return domainErr.ErrForbidden
	}

	// TODO: 检查工作空间是否有关联的资源（流程、工具、触发器、集群等）
	// 这里应该实现相关检查，确保工作空间是空的才能删除
	// 暂时先允许删除，但在实际环境中需要完善此检查

	slog.Warn("Workspace deletion without resource check - should be implemented",
		"workspaceId", workspaceID)

	// 删除工作空间
	err = s.workspaceRepo.Delete(ctx, workspaceID)
	if err != nil {
		slog.Error("Failed to delete workspace", "workspaceId", workspaceID, "error", err)
		return fmt.Errorf("delete workspace: %w", err)
	}

	slog.Info("Workspace deleted successfully", "workspaceId", workspaceID)
	return nil
}

func (s *workspaceServiceImpl) AddMember(ctx context.Context, workspaceID string, req *dto.AddWorkspaceMemberRequest, operatorID string) (*dto.WorkspaceMemberResponse, error) {
	slog.Info("Adding workspace member", "workspaceId", workspaceID, "accountId", req.AccountID, "role", req.Role)

	// 验证参数
	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		slog.Warn("Invalid workspace ID", "workspaceId", workspaceID, "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	// 查找工作空间
	workspace, err := s.workspaceRepo.FindByID(ctx, wsID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Workspace not found", "workspaceId", wsID)
			return nil, domainErr.ErrNotFound
		}
		slog.Error("Failed to find workspace", "workspaceId", wsID, "error", err)
		return nil, fmt.Errorf("find workspace: %w", err)
	}

	// 验证权限：只有所有者可以添加成员
	if workspace.OwnerID != operatorID {
		slog.Warn("User does not have permission to add member",
			"workspaceId", wsID, "operatorId", operatorID, "ownerId", workspace.OwnerID)
		return nil, domainErr.ErrForbidden
	}

	// 检查用户是否能添加自己为成员
	if req.AccountID == operatorID {
		slog.Warn("Cannot add owner as member", "accountId", req.AccountID)
		return nil, fmt.Errorf("owner cannot be added as member")
	}

	// 检查要添加的用户是否存在
	targetAccount, err := s.accountRepo.GetByID(ctx, req.AccountID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Target account not found", "accountId", req.AccountID)
			return nil, domainErr.ErrNotFound
		}
		slog.Error("Failed to get target account", "accountId", req.AccountID, "error", err)
		return nil, fmt.Errorf("get account: %w", err)
	}

	// TODO: 检查用户是否已经是成员
	// 这里需要实现 WorkspaceMemberRepository 来检查成员关系

	// TODO: 创建成员关系记录
	// 这里需要实现 WorkspaceMemberRepository 来保存成员关系

	slog.Info("Workspace member addition simulated - repository not implemented",
		"workspaceId", wsID, "accountId", req.AccountID, "role", req.Role)

	// 由于成员管理仓库未实现，返回一个模拟响应
	// TODO: 实现实际的成员添加逻辑后移除此模拟返回
	return &dto.WorkspaceMemberResponse{
		AccountID:   req.AccountID,
		Email:       targetAccount.Email.String(),
		DisplayName: targetAccount.DisplayName,
		Role:        req.Role,
		JoinedAt:    time.Now(),
	}, fmt.Errorf("workspace member management not fully implemented")
}

func (s *workspaceServiceImpl) RemoveMember(ctx context.Context, workspaceID string, memberID string, operatorID string) error {
	slog.Info("Removing workspace member", "workspaceId", workspaceID, "memberId", memberID, "operatorId", operatorID)

	// 验证参数
	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		slog.Warn("Invalid workspace ID", "workspaceId", workspaceID, "error", err)
		return domainErr.ErrInvalidInput
	}

	// 查找工作空间
	workspace, err := s.workspaceRepo.FindByID(ctx, wsID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Workspace not found", "workspaceId", wsID)
			return domainErr.ErrNotFound
		}
		slog.Error("Failed to find workspace", "workspaceId", wsID, "error", err)
		return fmt.Errorf("find workspace: %w", err)
	}

	// 验证权限：只有所有者可以移除成员
	if workspace.OwnerID != operatorID {
		slog.Warn("User does not have permission to remove member",
			"workspaceId", wsID, "operatorId", operatorID, "ownerId", workspace.OwnerID)
		return domainErr.ErrForbidden
	}

	// 检查是否尝试删除所有者
	if memberID == workspace.OwnerID {
		slog.Warn("Cannot remove workspace owner", "memberId", memberID)
		return fmt.Errorf("cannot remove workspace owner")
	}

	// 检查要移除的成员是否存在
	if _, err := s.accountRepo.GetByID(ctx, memberID); err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Member account not found", "memberId", memberID)
			return domainErr.ErrNotFound
		}
		slog.Error("Failed to get member account", "memberId", memberID, "error", err)
		return fmt.Errorf("get account: %w", err)
	}

	// TODO: 检查用户是否是成员
	// 这里需要实现 WorkspaceMemberRepository 来检查成员关系

	// TODO: 删除成员关系记录
	// 这里需要实现 WorkspaceMemberRepository 来删除成员关系

	slog.Info("Workspace member removal simulated - repository not implemented",
		"workspaceId", wsID, "memberId", memberID)

	// 由于成员管理仓库未实现，返回成功模拟
	// TODO: 实现实际的成员移除逻辑后移除此模拟返回
	return fmt.Errorf("workspace member management not fully implemented")
}

func (s *workspaceServiceImpl) ListMembers(ctx context.Context, workspaceID string, accountID string, page int, pageSize int) ([]dto.WorkspaceMemberResponse, error) {
	slog.Info("Listing workspace members", "workspaceId", workspaceID, "accountId", accountID, "page", page, "pageSize", pageSize)

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 验证参数
	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		slog.Warn("Invalid workspace ID", "workspaceId", workspaceID, "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	// 查找工作空间
	workspace, err := s.workspaceRepo.FindByID(ctx, wsID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Workspace not found", "workspaceId", wsID)
			return nil, domainErr.ErrNotFound
		}
		slog.Error("Failed to find workspace", "workspaceId", wsID, "error", err)
		return nil, fmt.Errorf("find workspace: %w", err)
	}

	// 验证权限：只有所有者或成员可以查看成员列表
	if workspace.OwnerID != accountID {
		// TODO: 实现成员权限检查，如果是成员也允许查看
		slog.Warn("User does not have permission to list members",
			"workspaceId", wsID, "userId", accountID, "ownerId", workspace.OwnerID)
		return nil, domainErr.ErrForbidden
	}

	// TODO: 实现成员列表查询
	// 这里需要实现 WorkspaceMemberRepository 来查询成员列表
	// 由于成员管理仓库未实现，返回模拟数据

	slog.Info("Workspace member listing simulated - repository not implemented",
		"workspaceId", wsID)

	// 模拟返回：只返回所有者信息
	ownerAccount, err := s.accountRepo.GetByID(ctx, workspace.OwnerID)
	if err != nil {
		slog.Error("Failed to get owner account", "ownerId", workspace.OwnerID, "error", err)
		return nil, fmt.Errorf("get owner account: %w", err)
	}

	// 创建模拟响应
	members := []dto.WorkspaceMemberResponse{
		{
			AccountID:   ownerAccount.ID,
			Email:       ownerAccount.Email.String(),
			DisplayName: ownerAccount.DisplayName,
			Role:        "owner",
			JoinedAt:    workspace.CreatedAt,
		},
	}

	slog.Info("Workspace members listed (simulated)", "workspaceId", wsID, "count", len(members))
	return members, fmt.Errorf("workspace member management not fully implemented")
}

// --- 私有方法 ---

// toWorkspaceResponse 将领域实体转换为DTO响应
func toWorkspaceResponse(workspace *entity.Workspace) *dto.WorkspaceResponse {
	return &dto.WorkspaceResponse{
		ID:          workspace.ID.String(),
		Name:        workspace.Name,
		Description: workspace.Description,
		OwnerID:     workspace.OwnerID,
		MemberCount: workspace.MemberCount,
		CreatedAt:   workspace.CreatedAt,
		UpdatedAt:   workspace.UpdatedAt,
	}
}
