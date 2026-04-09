package tool_repository

import (
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"

	"mengri-flow/internal/domain/entity"
)

// toModel 将实体转换为GORM模型
func toModel(e *entity.Tool) (*ToolModel, error) {
	configJSON, err := json.Marshal(e.Config)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	return &ToolModel{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		Type:        e.Type,
		Config:      datatypes.JSON(configJSON),
		ResourceID:  e.ResourceID,
		Version:     e.Version,
		Status:      string(e.Status),
		WorkspaceID: e.WorkspaceID,
		CreatedBy:   e.CreatedBy,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}, nil
}

// toEntity 将GORM模型转换为实体
func toEntity(m *ToolModel) (*entity.Tool, error) {
	var config map[string]interface{}
	if err := json.Unmarshal(m.Config, &config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &entity.Tool{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Type:        m.Type,
		Config:      config,
		ResourceID:  m.ResourceID,
		Version:     m.Version,
		Status:      entity.ToolStatus(m.Status),
		WorkspaceID: m.WorkspaceID,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}, nil
}
