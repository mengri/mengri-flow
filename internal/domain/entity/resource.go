package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ResourceType 资源类型
type ResourceType string

const (
	ResourceTypeHTTP     ResourceType = "http"
	ResourceTypeGRPC     ResourceType = "grpc"
	ResourceTypeMySQL    ResourceType = "mysql"
	ResourceTypePostgres ResourceType = "postgres"
)

// ResourceStatus 资源状态
type ResourceStatus string

const (
	ResourceStatusActive   ResourceStatus = "active"
	ResourceStatusInactive ResourceStatus = "inactive"
	ResourceStatusError    ResourceStatus = "error"
)

// Resource 表示一个资源
type Resource struct {
	ID          uuid.UUID
	Name        string
	Type        ResourceType
	Config      map[string]interface{}
	WorkspaceID uuid.UUID
	Status      ResourceStatus
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewResource 创建一个新的资源
func NewResource(name string, resourceType ResourceType, config map[string]interface{}, workspaceID uuid.UUID, description string) (*Resource, error) {
	if name == "" {
		return nil, fmt.Errorf("resource name cannot be empty")
	}

	if err := validateResourceType(resourceType); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Resource{
		ID:          uuid.New(),
		Name:        name,
		Type:        resourceType,
		Config:      config,
		WorkspaceID: workspaceID,
		Status:      ResourceStatusActive,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update 更新资源信息
func (r *Resource) Update(name string, config map[string]interface{}, description string) error {
	if name == "" {
		return fmt.Errorf("resource name cannot be empty")
	}

	r.Name = name
	r.Config = config
	r.Description = description
	r.UpdatedAt = time.Now()
	return nil
}

// UpdateStatus 更新资源状态
func (r *Resource) UpdateStatus(status ResourceStatus) error {
	if err := validateResourceStatus(status); err != nil {
		return err
	}
	r.Status = status
	r.UpdatedAt = time.Now()
	return nil
}

func validateResourceType(resourceType ResourceType) error {
	switch resourceType {
	case ResourceTypeHTTP, ResourceTypeGRPC, ResourceTypeMySQL, ResourceTypePostgres:
		return nil
	default:
		return fmt.Errorf("invalid resource type: %s", resourceType)
	}
}

func validateResourceStatus(status ResourceStatus) error {
	switch status {
	case ResourceStatusActive, ResourceStatusInactive, ResourceStatusError:
		return nil
	default:
		return fmt.Errorf("invalid resource status: %s", status)
	}
}
