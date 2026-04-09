package service

import (
	"context"
	"fmt"
	"time"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/domain/entity"
	"mengri-flow/internal/domain/repository"
	"mengri-flow/internal/infra/plugin"

	"github.com/google/uuid"
)

type ResourceServiceImpl struct {
	resourceRepo   repository.ResourceRepository `autowired:""`
	pluginRegistry *plugin.Registry
}

func (s *ResourceServiceImpl) CreateResource(ctx context.Context, req *dto.CreateResourceRequest) (*entity.Resource, error) {
	resourceType := entity.ResourceType(req.Type)
	plugin, ok := s.pluginRegistry.GetResource(string(resourceType))
	if !ok {
		return nil, fmt.Errorf("unsupported resource type: %s", req.Type)
	}

	if err := plugin.TestConnection(ctx, req.Config); err != nil {
		return nil, fmt.Errorf("connection test failed: %w", err)
	}

	resource := &entity.Resource{
		ID:          uuid.New(),
		Name:        req.Name,
		Type:        resourceType,
		Config:      req.Config,
		WorkspaceID: uuid.MustParse(req.WorkspaceID),
		Status:      entity.ResourceStatusActive,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.resourceRepo.Create(ctx, resource); err != nil {
		return nil, err
	}

	return resource, nil
}

func (s *ResourceServiceImpl) UpdateResource(ctx context.Context, id string, req *dto.UpdateResourceRequest) (*entity.Resource, error) {
	resource, err := s.resourceRepo.FindByID(ctx, uuid.MustParse(id))
	if err != nil {
		return nil, err
	}

	if req.Config != nil {
		plugin, ok := s.pluginRegistry.GetResource(string(resource.Type))
		if !ok {
			return nil, fmt.Errorf("unsupported resource type: %s", resource.Type)
		}

		if err := plugin.TestConnection(ctx, req.Config); err != nil {
			return nil, fmt.Errorf("connection test failed: %w", err)
		}

		resource.Config = req.Config
	}

	if req.Name != "" {
		resource.Name = req.Name
	}
	if req.Description != "" {
		resource.Description = req.Description
	}
	resource.UpdatedAt = time.Now()

	if err := s.resourceRepo.Update(ctx, resource); err != nil {
		return nil, err
	}

	return resource, nil
}

func (s *ResourceServiceImpl) DeleteResource(ctx context.Context, id string) error {
	return s.resourceRepo.Delete(ctx, uuid.MustParse(id))
}

func (s *ResourceServiceImpl) GetResource(ctx context.Context, id string) (*entity.Resource, error) {
	return s.resourceRepo.FindByID(ctx, uuid.MustParse(id))
}

func (s *ResourceServiceImpl) ListResources(ctx context.Context, workspaceID string, resourceType string, page int, pageSize int) ([]*entity.Resource, int64, error) {
	if resourceType != "" {
		resources, err := s.resourceRepo.ListByType(ctx, entity.ResourceType(resourceType))
		if err != nil {
			return nil, 0, err
		}
		// TODO: Implement pagination for ListByType
		return resources, int64(len(resources)), nil
	}
	resources, err := s.resourceRepo.ListByWorkspace(ctx, uuid.MustParse(workspaceID))
	if err != nil {
		return nil, 0, err
	}
	// TODO: Implement pagination for ListByWorkspace
	return resources, int64(len(resources)), nil
}

func (s *ResourceServiceImpl) TestConnection(ctx context.Context, req *dto.TestConnectionRequest) error {
	resourceType := entity.ResourceType(req.Type)
	plugin, ok := s.pluginRegistry.GetResource(string(resourceType))
	if !ok {
		return fmt.Errorf("unsupported resource type: %s", req.Type)
	}

	return plugin.TestConnection(ctx, req.Config)
}
