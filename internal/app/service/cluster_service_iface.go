package service

import (
	"context"

	"mengri-flow/internal/app/dto"
)

type IClusterService interface {
	CreateCluster(ctx context.Context, req *dto.CreateClusterRequest) (*dto.ClusterResponse, error)
	ListClusters(ctx context.Context, environmentID string, page int, pageSize int) (*dto.ListClustersResponse, error)
	GetClusterDetail(ctx context.Context, id string) (*dto.ClusterDetailResponse, error)
	UpdateCluster(ctx context.Context, id string, req *dto.UpdateClusterRequest) (*dto.ClusterResponse, error)
	DeleteCluster(ctx context.Context, id string) error
	TestEtcdConnection(ctx context.Context, req *dto.TestEtcdConnectionRequest) (*dto.TestEtcdConnectionResponse, error)
}

var _ IClusterService = (*ClusterService)(nil)
