package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ToolStatus 工具状态
type ToolStatus string

const (
	ToolStatusDraft     ToolStatus = "draft"
	ToolStatusPublished ToolStatus = "published"
	ToolStatusDeprecated ToolStatus = "deprecated"
)

// Tool 表示一个工具
type Tool struct {
	ID          uuid.UUID
	Name        string
	Description string
	Type        string
	Config      map[string]interface{}
	ResourceID  uuid.UUID
	Version     int
	Status      ToolStatus
	WorkspaceID uuid.UUID
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewTool 创建一个新的工具
func NewTool(name string, toolType string, config map[string]interface{}, resourceID, workspaceID uuid.UUID, createdBy string) (*Tool, error) {
	if name == "" {
		return nil, fmt.Errorf("tool name cannot be empty")
	}

	now := time.Now()
	return &Tool{
		ID:          uuid.New(),
		Name:        name,
		Type:        toolType,
		Config:      config,
		ResourceID:  resourceID,
		Version:     1,
		Status:      ToolStatusDraft,
		WorkspaceID: workspaceID,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update 更新工具信息
func (t *Tool) Update(name string, description string, config map[string]interface{}) error {
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	t.Name = name
	t.Description = description
	t.Config = config
	t.Version++
	t.UpdatedAt = time.Now()
	return nil
}

// Publish 发布工具
func (t *Tool) Publish() error {
	if t.Status != ToolStatusDraft {
		return fmt.Errorf("only draft tools can be published")
	}
	t.Status = ToolStatusPublished
	t.Version++
	t.UpdatedAt = time.Now()
	return nil
}

// Deprecate 废弃工具
func (t *Tool) Deprecate() error {
	if t.Status != ToolStatusPublished {
		return fmt.Errorf("only published tools can be deprecated")
	}
	t.Status = ToolStatusDeprecated
	t.UpdatedAt = time.Now()
	return nil
}
