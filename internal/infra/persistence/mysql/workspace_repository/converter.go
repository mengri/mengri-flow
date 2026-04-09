package workspaceRepository

import (
	"github.com/google/uuid"
	"mengri-flow/internal/domain/entity"
)

// toModel 将领域实体转换为GORM模型
func toModel(workspace *entity.Workspace) *WorkspaceModel {
	return &WorkspaceModel{
		ID:          workspace.ID.String(),
		Name:        workspace.Name,
		Description: workspace.Description,
		OwnerID:     workspace.OwnerID,
		MemberCount: workspace.MemberCount,
		CreatedAt:   workspace.CreatedAt,
		UpdatedAt:   workspace.UpdatedAt,
	}
}

// toEntity 将GORM模型转换为领域实体
func toEntity(model *WorkspaceModel) (*entity.Workspace, error) {
	id, err := uuid.Parse(model.ID)
	if err != nil {
		return nil, err
	}

	return &entity.Workspace{
		ID:          id,
		Name:        model.Name,
		Description: model.Description,
		OwnerID:     model.OwnerID,
		MemberCount: model.MemberCount,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}, nil
}