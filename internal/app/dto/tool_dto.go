package dto

import (
	"time"
)

type CreateToolRequest struct {
	Name        string                 `json:"name" binding:"required,max=100"`
	Type        string                 `json:"type" binding:"required"`
	Config      map[string]interface{} `json:"config" binding:"required"`
	WorkspaceID string                 `json:"workspaceId" binding:"required"`
	Description string                 `json:"description" binding:"max=500"`
}

type UpdateToolRequest struct {
	Name        string                 `json:"name" binding:"omitempty,max=100"`
	Config      map[string]interface{} `json:"config"`
	Description string                 `json:"description" binding:"omitempty,max=500"`
}

type ToolResponse struct {
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

type TestToolRequest struct {
	ToolID string                 `json:"toolId" binding:"required"`
	Input  map[string]interface{} `json:"input" binding:"required"`
}

type ImportToolsRequest struct {
	ResourceID string `json:"resourceId" binding:"required"`
}

type PublishToolRequest struct {
	ToolID string `json:"toolId" binding:"required"`
}

type ListToolsRequest struct {
	WorkspaceID string `json:"workspaceId" binding:"required"`
	Type        string `json:"type" binding:"omitempty"`
	Status      string `json:"status" binding:"omitempty"`
	Page        int    `json:"page" binding:"min=1"`
	PageSize    int    `json:"pageSize" binding:"min=1,max=100"`
}

type ListToolsResponse struct {
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
	List     []ToolResponse `json:"list"`
}

// ToolVersionResponse 工具版本响应
type ToolVersionResponse struct {
	Version   int       `json:"version"`
	Comment   string    `json:"comment"`
	CreatedBy string    `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
}
