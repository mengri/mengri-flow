package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/domain/entity"
	"mengri-flow/internal/domain/repository"
	"mengri-flow/internal/infra/plugin"
)

type ResourceService struct {
	resourceRepo   repository.ResourceRepository
	pluginRegistry *plugin.Registry
}

func NewResourceService(repo repository.ResourceRepository, registry *plugin.Registry) *ResourceService {
	return &ResourceService{
		resourceRepo:   repo,
		pluginRegistry: registry,
	}
}

func (s *ResourceService) CreateResource(ctx context.Context, req *dto.CreateResourceRequest) (*entity.Resource, error) {
	plugin, err := s.pluginRegistry.GetResource(req.Type)
	if err != nil {
		return nil, fmt.Errorf("unsupported resource type: %s", req.Type)
	}

	if err := plugin.TestConnection(ctx, req.Config); err != nil {
		return nil, fmt.Errorf("connection test failed: %w", err)
	}

	resource := &entity.Resource{
		ID:          uuid.New(),
		Name:        req.Name,
		Type:        req.Type,
		Config:      req.Config,
		WorkspaceID: uuid.MustParse(req.WorkspaceID),
		Status:      "active",
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.resourceRepo.Create(ctx, resource); err != nil {
		return nil, err
	}

	return resource, nil
}

func (s *ResourceService) UpdateResource(ctx context.Context, id string, req *dto.UpdateResourceRequest) (*entity.Resource, error) {
	resource, err := s.resourceRepo.FindByID(ctx, uuid.MustParse(id))
	if err != nil {
		return nil, err
	}

	if req.Config != nil {
		plugin, err := s.pluginRegistry.GetResource(resource.Type)
		if err != nil {
			return nil, err
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

func (s *ResourceService) DeleteResource(ctx context.Context, id string) error {
	return s.resourceRepo.Delete(ctx, uuid.MustParse(id))
}

func (s *ResourceService) GetResource(ctx context.Context, id string) (*entity.Resource, error) {
	return s.resourceRepo.FindByID(ctx, uuid.MustParse(id))
}

func (s *ResourceService) ListResources(ctx context.Context, workspaceID string, resourceType string, page int, pageSize int) ([]*entity.Resource, int64, error) {
	if resourceType != "" {
		return s.resourceRepo.FindByType(ctx, uuid.MustParse(workspaceID), resourceType, page, pageSize)
	}
	return s.resourceRepo.FindByWorkspace(ctx, uuid.MustParse(workspaceID), page, pageSize)
}

func (s *ResourceService) TestConnection(ctx context.Context, req *dto.TestConnectionRequest) error {
	plugin, err := s.pluginRegistry.GetResource(req.Type)
	if err != nil {
		return fmt.Errorf("unsupported resource type: %s", req.Type)
	}

	return plugin.TestConnection(ctx, req.Config)
}
