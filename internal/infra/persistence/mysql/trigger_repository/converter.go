package triggerRepository

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"mengri-flow/internal/domain/entity"
)

// toModel 将领域实体转换为GORM模型
func toModel(trigger *entity.Trigger) *TriggerModel {
	configJSON, _ := json.Marshal(trigger.Config)
	inputMappingJSON, _ := json.Marshal(trigger.InputMapping)
	outputMappingJSON, _ := json.Marshal(trigger.OutputMapping)
	errorHandlingJSON, _ := json.Marshal(trigger.ErrorHandling)

	return &TriggerModel{
		ID:             trigger.ID.String(),
		Name:           trigger.Name,
		Type:           string(trigger.Type),
		FlowID:         trigger.FlowID.String(),
		FlowVersion:    trigger.FlowVersion,
		ClusterID:      trigger.ClusterID.String(),
		Config:         string(configJSON),
		InputMapping:   string(inputMappingJSON),
		OutputMapping:  string(outputMappingJSON),
		ErrorHandling:  string(errorHandlingJSON),
		WorkspaceID:    trigger.WorkspaceID.String(),
		Status:         string(trigger.Status),
		LastExecutedAt: trigger.LastExecutedAt,
		CreatedAt:      trigger.CreatedAt,
		UpdatedAt:      trigger.UpdatedAt,
	}
}

// toEntity 将GORM模型转换为领域实体
func toEntity(model *TriggerModel) (*entity.Trigger, error) {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(model.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	var inputMapping map[string]string
	if model.InputMapping != "" && model.InputMapping != "null" {
		if err := json.Unmarshal([]byte(model.InputMapping), &inputMapping); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input mapping: %w", err)
		}
	}

	var outputMapping map[string]string
	if model.OutputMapping != "" && model.OutputMapping != "null" {
		if err := json.Unmarshal([]byte(model.OutputMapping), &outputMapping); err != nil {
			return nil, fmt.Errorf("failed to unmarshal output mapping: %w", err)
		}
	}

	var errorHandling entity.ErrorHandling
	if model.ErrorHandling != "" && model.ErrorHandling != "null" {
		if err := json.Unmarshal([]byte(model.ErrorHandling), &errorHandling); err != nil {
			return nil, fmt.Errorf("failed to unmarshal error handling: %w", err)
		}
	}

	var clusterID uuid.UUID
	if model.ClusterID != "" {
		clusterID = uuid.MustParse(model.ClusterID)
	}

	var workspaceID uuid.UUID
	if model.WorkspaceID != "" {
		workspaceID = uuid.MustParse(model.WorkspaceID)
	}

	return &entity.Trigger{
		ID:             uuid.MustParse(model.ID),
		Name:           model.Name,
		Type:           entity.TriggerType(model.Type),
		FlowID:         uuid.MustParse(model.FlowID),
		FlowVersion:    model.FlowVersion,
		ClusterID:      clusterID,
		Config:         config,
		InputMapping:   inputMapping,
		OutputMapping:  outputMapping,
		ErrorHandling:  errorHandling,
		WorkspaceID:    workspaceID,
		Status:         entity.TriggerStatus(model.Status),
		LastExecutedAt: model.LastExecutedAt,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}, nil
}