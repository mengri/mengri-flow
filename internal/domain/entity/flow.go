package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FlowStatus 流程状态
type FlowStatus string

const (
	FlowStatusDraft     FlowStatus = "draft"
	FlowStatusActive    FlowStatus = "active"
	FlowStatusInactive  FlowStatus = "inactive"
)

// Flow 表示一个流程
type Flow struct {
	ID          uuid.UUID
	Name        string
	Description string
	CanvasData  map[string]interface{}
	Status      FlowStatus
	Version     int
	WorkspaceID uuid.UUID
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewFlow 创建一个新的流程
func NewFlow(name string, canvasData map[string]interface{}, workspaceID uuid.UUID, createdBy string) (*Flow, error) {
	if name == "" {
		return nil, fmt.Errorf("flow name cannot be empty")
	}

	now := time.Now()
	return &Flow{
		ID:          uuid.New(),
		Name:        name,
		CanvasData:  canvasData,
		Status:      FlowStatusDraft,
		Version:     1,
		WorkspaceID: workspaceID,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update 更新流程信息
func (f *Flow) Update(name string, description string, canvasData map[string]interface{}) error {
	if name == "" {
		return fmt.Errorf("flow name cannot be empty")
	}

	f.Name = name
	f.Description = description
	f.CanvasData = canvasData
	f.Version++
	f.UpdatedAt = time.Now()
	return nil
}

// UpdateStatus 更新流程状态
func (f *Flow) UpdateStatus(status FlowStatus) error {
	if err := validateFlowStatus(status); err != nil {
		return err
	}
	f.Status = status
	f.UpdatedAt = time.Now()
	return nil
}

// Publish 发布流程
func (f *Flow) Publish() error {
	if f.Status != FlowStatusDraft && f.Status != FlowStatusInactive {
		return fmt.Errorf("only draft or inactive flows can be published")
	}
	f.Status = FlowStatusActive
	f.Version++
	f.UpdatedAt = time.Now()
	return nil
}

func validateFlowStatus(status FlowStatus) error {
	switch status {
	case FlowStatusDraft, FlowStatusActive, FlowStatusInactive:
		return nil
	default:
		return fmt.Errorf("invalid flow status: %s", status)
	}
}
