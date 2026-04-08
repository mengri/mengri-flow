package service

import (
	"context"

	"mengri-flow/internal/app/dto"
)

type IFlowService interface {
	CreateFlow(ctx context.Context, req *dto.CreateFlowRequest, creatorID string) (*dto.FlowResponse, error)
	ListFlows(ctx context.Context, req *dto.ListFlowsRequest) (*dto.ListFlowsResponse, error)
	GetFlow(ctx context.Context, id string) (*dto.FlowResponse, error)
	UpdateFlow(ctx context.Context, id string, req *dto.UpdateFlowRequest, operatorID string) (*dto.FlowResponse, error)
	DeleteFlow(ctx context.Context, id string, operatorID string) error
	TestFlow(ctx context.Context, req *dto.TestFlowRequest) error
	PublishFlow(ctx context.Context, flowID string, clusterID string, operatorID string) error
	ListVersions(ctx context.Context, flowID string) ([]int, error)
	RollbackVersion(ctx context.Context, flowID string, version int, operatorID string) error
}

var _ IFlowService = (*FlowService)(nil)
