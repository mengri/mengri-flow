package dto

import (
	"time"
)

type CreateFlowRequest struct {
	Name        string                 `json:"name" binding:"required,max=100"`
	Description string                 `json:"description" binding:"max=500"`
	WorkspaceID string                 `json:"workspaceId" binding:"required"`
	Config      map[string]interface{} `json:"config" binding:"required"`
}

type UpdateFlowRequest struct {
	Name        string                 `json:"name" binding:"omitempty,max=100"`
	Description string                 `json:"description" binding:"omitempty,max=500"`
	Config      map[string]interface{} `json:"config"`
}

type FlowResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	WorkspaceID string                 `json:"workspaceId"`
	Config      map[string]interface{} `json:"config"`
	Version     int                    `json:"version"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

type TestFlowRequest struct {
	FlowID string                 `json:"flowId" binding:"required"`
	Input  map[string]interface{} `json:"input" binding:"required"`
}

type PublishFlowRequest struct {
	FlowID    string `json:"flowId" binding:"required"`
	ClusterID string `json:"clusterId" binding:"required"`
}

type RollbackFlowRequest struct {
	FlowID  string `json:"flowId" binding:"required"`
	Version int    `json:"version" binding:"required,min=1"`
}

type ListFlowsRequest struct {
	WorkspaceID string `json:"workspaceId" binding:"required"`
	Status      string `json:"status" binding:"omitempty"`
	Page        int    `json:"page" binding:"min=1"`
	PageSize    int    `json:"pageSize" binding:"min=1,max=100"`
}

type ListFlowsResponse struct {
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
	List     []FlowResponse `json:"list"`
}
