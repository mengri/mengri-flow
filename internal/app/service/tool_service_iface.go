package service

import (
	"context"

	"mengri-flow/internal/app/dto"
	"mengri-flow/pkg/autowire"
)

type IToolService interface {
	CreateTool(ctx context.Context, req *dto.CreateToolRequest) (*dto.ToolResponse, error)
	ListTools(ctx context.Context, req *dto.ListToolsRequest) (*dto.ListToolsResponse, error)
	GetTool(ctx context.Context, id string) (*dto.ToolResponse, error)
	UpdateTool(ctx context.Context, id string, req *dto.UpdateToolRequest) (*dto.ToolResponse, error)
	TestTool(ctx context.Context, req *dto.TestToolRequest) error
	ImportTools(ctx context.Context, req *dto.ImportToolsRequest) error
	PublishTool(ctx context.Context, toolID string) error
	DeprecateTool(ctx context.Context, toolID string) error
	ListVersions(ctx context.Context, toolID string) ([]string, error)
}

var _ IToolService = (*toolServiceImpl)(nil)

func init() {
	autowire.Auto(func() IToolService {
		return new(toolServiceImpl)
	})
}
