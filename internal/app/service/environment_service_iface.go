package service

import (
	"context"

	"mengri-flow/internal/app/dto"
	"mengri-flow/pkg/autowire"
)

type IEnvironmentService interface {
	CreateEnvironment(ctx context.Context, req *dto.CreateEnvironmentRequest) (*dto.EnvironmentResponse, error)
	ListEnvironments(ctx context.Context, page int, pageSize int) (*dto.ListEnvironmentsResponse, error)
	GetEnvironment(ctx context.Context, id string) (*dto.EnvironmentResponse, error)
	UpdateEnvironment(ctx context.Context, id string, req *dto.UpdateEnvironmentRequest) (*dto.EnvironmentResponse, error)
	DeleteEnvironment(ctx context.Context, id string) error
}

var _ IEnvironmentService = (*environmentServiceImpl)(nil)

func init() {
	autowire.Auto(func() IEnvironmentService {
		return &environmentServiceImpl{}
	})
}
