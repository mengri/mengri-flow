package resourceRepository

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"mengri-flow/internal/domain/entity"
)

// toModel 将领域实体转换为GORM模型
func toModel(resource *entity.Resource) *ResourceModel {
	configJSON, _ := json.Marshal(resource.Config)

	return &ResourceModel{
		ID:          resource.ID.String(),
		Name:        resource.Name,
		Type:        string(resource.Type),
		Config:      string(configJSON),
		WorkspaceID: resource.WorkspaceID.String(),
		Status:      string(resource.Status),
		Description: resource.Description,
		CreatedAt:   resource.CreatedAt.Unix(),
		UpdatedAt:   resource.UpdatedAt.Unix(),
	}
}

// toEntity 将GORM模型转换为领域实体
func toEntity(model *ResourceModel) (*entity.Resource, error) {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(model.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &entity.Resource{
		ID:          uuid.MustParse(model.ID),
		Name:        model.Name,
		Type:        entity.ResourceType(model.Type),
		Config:      config,
		WorkspaceID: uuid.MustParse(model.WorkspaceID),
		Status:      entity.ResourceStatus(model.Status),
		Description: model.Description,
		CreatedAt:   time.Unix(model.CreatedAt, 0),
		UpdatedAt:   time.Unix(model.UpdatedAt, 0),
	}, nil
}
