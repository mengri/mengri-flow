package service

import (
	"context"

	"mengri-flow/internal/app/dto"
	"mengri-flow/pkg/autowire"
)

type IWorkspaceService interface {
	CreateWorkspace(ctx context.Context, req *dto.CreateWorkspaceRequest, ownerID string) (*dto.WorkspaceResponse, error)
	ListWorkspaces(ctx context.Context, accountID string, page int, pageSize int) (*dto.ListWorkspacesResponse, error)
	GetWorkspace(ctx context.Context, id string, accountID string) (*dto.WorkspaceResponse, error)
	UpdateWorkspace(ctx context.Context, id string, req *dto.UpdateWorkspaceRequest, accountID string) (*dto.WorkspaceResponse, error)
	DeleteWorkspace(ctx context.Context, id string, accountID string) error
	AddMember(ctx context.Context, workspaceID string, req *dto.AddWorkspaceMemberRequest, operatorID string) (*dto.WorkspaceMemberResponse, error)
	RemoveMember(ctx context.Context, workspaceID string, memberID string, operatorID string) error
	ListMembers(ctx context.Context, workspaceID string, accountID string, page int, pageSize int) ([]dto.WorkspaceMemberResponse, int64, error)
}

var _ IWorkspaceService = (*workspaceServiceImpl)(nil)

func init() {
	autowire.Auto(func() IWorkspaceService {
		return new(workspaceServiceImpl)
	})
}
