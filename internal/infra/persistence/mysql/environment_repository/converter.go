package environmentRepository

import (
	"github.com/google/uuid"
	"mengri-flow/internal/domain/entity"
)

// toModel 将领域实体转换为GORM模型
func toModel(env *entity.Environment) *EnvironmentModel {
	return &EnvironmentModel{
		ID:          env.ID.String(),
		Name:        env.Name,
		Key:         env.Key,
		Description: env.Description,
		Color:       env.Color,
		CreatedAt:   env.CreatedAt,
		UpdatedAt:   env.UpdatedAt,
	}
}

// toEntity 将GORM模型转换为领域实体
func toEntity(model *EnvironmentModel) *entity.Environment {
	return &entity.Environment{
		ID:          uuid.MustParse(model.ID),
		Name:        model.Name,
		Key:         model.Key,
		Description: model.Description,
		Color:       model.Color,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}
