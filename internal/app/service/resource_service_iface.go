package service

import (
	"context"
	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/domain/entity"
	"mengri-flow/internal/infra/plugin"
	"mengri-flow/pkg/autowire"
)

// IResourceService 资源服务接口
type IResourceService interface {
	CreateResource(ctx context.Context, req *dto.CreateResourceRequest) (*entity.Resource, error)
	UpdateResource(ctx context.Context, id string, req *dto.UpdateResourceRequest) (*entity.Resource, error)
	DeleteResource(ctx context.Context, id string) error
	GetResource(ctx context.Context, id string) (*entity.Resource, error)
	ListResources(ctx context.Context, workspaceID string, resourceType string, page int, pageSize int) ([]*entity.Resource, int64, error)
	TestConnection(ctx context.Context, req *dto.TestConnectionRequest) error
}

func init() {
	autowire.Auto(func() IResourceService {
		return &ResourceServiceImpl{pluginRegistry: plugin.GlobalRegistry()}
	})
}
