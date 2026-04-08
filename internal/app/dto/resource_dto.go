package dto

import (
	"time"
)

type CreateResourceRequest struct {
	Name        string                 `json:"name" binding:"required,max=100"`
	Type        string                 `json:"type" binding:"required"`
	Config      map[string]interface{} `json:"config" binding:"required"`
	WorkspaceID string                 `json:"workspaceId" binding:"required"`
	Description string                 `json:"description" binding:"max=500"`
}

type UpdateResourceRequest struct {
	Name        string                 `json:"name" binding:"omitempty,max=100"`
	Config      map[string]interface{} `json:"config"`
	Description string                 `json:"description" binding:"omitempty,max=500"`
}

type ResourceResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Config      map[string]interface{} `json:"config"`
	Status      string                 `json:"status"`
	WorkspaceID string                 `json:"workspaceId"`
	Description string                 `json:"description"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

type TestConnectionRequest struct {
	Type   string                 `json:"type" binding:"required"`
	Config map[string]interface{} `json:"config" binding:"required"`
}

type ListResourcesRequest struct {
	WorkspaceID string `json:"workspaceId" binding:"required"`
	Type        string `json:"type" binding:"omitempty"`
	Page        int    `json:"page" binding:"min=1"`
	PageSize    int    `json:"pageSize" binding:"min=1,max=100"`
}

type ListResourcesResponse struct {
	Total   int64              `json:"total"`
	Page    int                `json:"page"`
	PageSize int               `json:"pageSize"`
	List    []ResourceResponse `json:"list"`
}
