package workspaceMemberRepository

import (
	"mengri-flow/internal/domain/entity"
)

// toModel 将领域实体转换为 GORM 模型
func toModel(member *entity.WorkspaceMember) *WorkspaceMemberModel {
	return &WorkspaceMemberModel{
		WorkspaceID: member.WorkspaceID,
		AccountID:   member.AccountID,
		Role:        string(member.Role),
		JoinedAt:    member.JoinedAt,
	}
}

// toEntity 将 GORM 模型转换为领域实体
func toEntity(model *WorkspaceMemberModel) *entity.WorkspaceMember {
	return &entity.WorkspaceMember{
		WorkspaceID: model.WorkspaceID,
		AccountID:   model.AccountID,
		Role:        entity.MemberRole(model.Role),
		JoinedAt:    model.JoinedAt,
	}
}
