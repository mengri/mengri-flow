package flow_repository

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"mengri-flow/internal/domain/entity"
)

// toModel 将实体转换为GORM模型
func toModel(e *entity.Flow) (*FlowModel, error) {
	canvasDataJSON, err := json.Marshal(e.CanvasData)
	if err != nil {
		return nil, fmt.Errorf("marshal canvas data: %w", err)
	}

	return &FlowModel{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		CanvasData:  datatypes.JSON(canvasDataJSON),
		Status:      string(e.Status),
		Version:     e.Version,
		WorkspaceID: e.WorkspaceID,
		CreatedBy:   e.CreatedBy,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}, nil
}

// toVersionModel 将实体转换为版本模型
func toVersionModel(e *entity.Flow) (*FlowVersionModel, error) {
	canvasDataJSON, err := json.Marshal(e.CanvasData)
	if err != nil {
		return nil, fmt.Errorf("marshal canvas data: %w", err)
	}

	return &FlowVersionModel{
		ID:          uuid.New(), // 版本记录使用新ID
		FlowID:      e.ID,
		Name:        e.Name,
		Description: e.Description,
		CanvasData:  datatypes.JSON(canvasDataJSON),
		Status:      string(e.Status),
		Version:     e.Version,
		WorkspaceID: e.WorkspaceID,
		CreatedBy:   e.CreatedBy,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}, nil
}

// toEntity 将GORM模型转换为实体
func toEntity(m *FlowModel) (*entity.Flow, error) {
	var canvasData map[string]interface{}
	if err := json.Unmarshal(m.CanvasData, &canvasData); err != nil {
		return nil, fmt.Errorf("unmarshal canvas data: %w", err)
	}

	return &entity.Flow{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		CanvasData:  canvasData,
		Status:      entity.FlowStatus(m.Status),
		Version:     m.Version,
		WorkspaceID: m.WorkspaceID,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}, nil
}

// toEntityFromVersion 将版本模型转换为实体
func toEntityFromVersion(m *FlowVersionModel) (*entity.Flow, error) {
	var canvasData map[string]interface{}
	if err := json.Unmarshal(m.CanvasData, &canvasData); err != nil {
		return nil, fmt.Errorf("unmarshal canvas data: %w", err)
	}

	return &entity.Flow{
		ID:          m.FlowID,
		Name:        m.Name,
		Description: m.Description,
		CanvasData:  canvasData,
		Status:      entity.FlowStatus(m.Status),
		Version:     m.Version,
		WorkspaceID: m.WorkspaceID,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}, nil
}