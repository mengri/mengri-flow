package dto

import (
	"time"
)

type CreateTriggerRequest struct {
	Name          string                 `json:"name" binding:"required,max=100"`
	Type          string                 `json:"type" binding:"required"`
	Config        map[string]interface{} `json:"config" binding:"required"`
	FlowID        string                 `json:"flowId" binding:"required"`
	FlowVersion   int                    `json:"flowVersion"`
	ClusterID     string                 `json:"clusterId"`
	InputMapping  map[string]string      `json:"inputMapping"`
	OutputMapping map[string]string      `json:"outputMapping"`
	ErrorHandling *ErrorHandlingConfig   `json:"errorHandling"`
	WorkspaceID   string                 `json:"workspaceId"`
	Description   string                 `json:"description" binding:"max=500"`
}

type ErrorHandlingConfig struct {
	Strategy         string                  `json:"strategy"`
	CustomErrorFormat  map[string]interface{}  `json:"customErrorFormat,omitempty"`
	RetryOnFailure   bool                    `json:"retryOnFailure"`
}

type UpdateTriggerRequest struct {
	Name          string                 `json:"name" binding:"omitempty,max=100"`
	Config        map[string]interface{} `json:"config"`
	FlowVersion   *int                    `json:"flowVersion,omitempty"`
	ClusterID     *string                 `json:"clusterId,omitempty"`
	InputMapping  *map[string]string      `json:"inputMapping,omitempty"`
	OutputMapping *map[string]string      `json:"outputMapping,omitempty"`
	ErrorHandling *ErrorHandlingConfig   `json:"errorHandling,omitempty"`
	Description   string                 `json:"description" binding:"omitempty,max=500"`
}

type TriggerResponse struct {
	ID            string                  `json:"id"`
	Name          string                  `json:"name"`
	Type          string                  `json:"type"`
	Config        map[string]interface{}  `json:"config"`
	FlowID        string                  `json:"flowId"`
	FlowVersion   int                     `json:"flowVersion"`
	ClusterID     string                  `json:"clusterId"`
	InputMapping  map[string]string       `json:"inputMapping"`
	OutputMapping map[string]string       `json:"outputMapping"`
	ErrorHandling *ErrorHandlingResponse  `json:"errorHandling"`
	WorkspaceID   string                  `json:"workspaceId"`
	Status        string                  `json:"status"`
	Description   string                  `json:"description"`
	CreatedAt     time.Time               `json:"createdAt"`
	UpdatedAt     time.Time               `json:"updatedAt"`
}

type ErrorHandlingResponse struct {
	Strategy          string                 `json:"strategy"`
	CustomErrorFormat map[string]interface{} `json:"customErrorFormat,omitempty"`
	RetryOnFailure    bool                   `json:"retryOnFailure"`
}

type EnableTriggerRequest struct {
	TriggerID string `json:"triggerId" binding:"required"`
}

type PublishTriggerRequest struct {
	TriggerID string `json:"triggerId" binding:"required"`
	ClusterID string `json:"clusterId" binding:"required"`
}

type ListTriggersRequest struct {
	WorkspaceID string `json:"workspaceId" binding:"omitempty"`
	FlowID     string `json:"flowId" binding:"omitempty"`
	Status     string `json:"status" binding:"omitempty"`
	Page       int    `json:"page" binding:"min=1"`
	PageSize   int    `json:"pageSize" binding:"min=1,max=100"`
}

type ListTriggersResponse struct {
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
	List     []TriggerResponse `json:"list"`
}
