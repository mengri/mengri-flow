package dto

import (
	"time"
)

type CreateWorkspaceRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description" binding:"max=500"`
}

type UpdateWorkspaceRequest struct {
	Name        string `json:"name" binding:"omitempty,max=100"`
	Description string `json:"description" binding:"omitempty,max=500"`
}

type WorkspaceResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"ownerId"`
	MemberCount int       `json:"memberCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type AddWorkspaceMemberRequest struct {
	AccountID string `json:"accountId" binding:"required"`
	Role      string `json:"role" binding:"required,oneof=member admin"`
}

type WorkspaceMemberResponse struct {
	AccountID   string    `json:"accountId"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joinedAt"`
}

type ListWorkspacesRequest struct {
	Page     int `json:"page" binding:"min=1"`
	PageSize int `json:"pageSize" binding:"min=1,max=100"`
}

type ListWorkspacesResponse struct {
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
	List     []WorkspaceResponse `json:"list"`
}

type ListWorkspaceMembersResponse struct {
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"pageSize"`
	List     []WorkspaceMemberResponse `json:"list"`
}
