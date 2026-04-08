package dto

import (
	"time"
)

type CreateTriggerRequest struct {
	Name        string                 `json:"name" binding:"required,max=100"`
	Type        string                 `json:"type" binding:"required"`
	Config      map[string]interface{} `json:"config" binding:"required"`
	FlowID      string                 `json:"flowId" binding:"required"`
	Description string                 `json:"description" binding:"max=500"`
}

type UpdateTriggerRequest struct {
	Name        string                 `json:"name" binding:"omitempty,max=100"`
	Config      map[string]interface{} `json:"config"`
	Description string                 `json:"description" binding:"omitempty,max=500"`
}

type TriggerResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Config      map[string]interface{} `json:"config"`
	FlowID      string                 `json:"flowId"`
	Status      string                 `json:"status"`
	Description string                 `json:"description"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

type EnableTriggerRequest struct {
	TriggerID string `json:"triggerId" binding:"required"`
}

type PublishTriggerRequest struct {
	TriggerID string `json:"triggerId" binding:"required"`
	ClusterID string `json:"clusterId" binding:"required"`
}

type ListTriggersRequest struct {
	FlowID string `json:"flowId" binding:"omitempty"`
	Status string `json:"status" binding:"omitempty"`
	Page   int    `json:"page" binding:"min=1"`
	PageSize int  `json:"pageSize" binding:"min=1,max=100"`
}

type ListTriggersResponse struct {
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
	List     []TriggerResponse `json:"list"`
}
