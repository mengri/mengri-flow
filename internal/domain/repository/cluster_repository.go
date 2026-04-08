package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"mengri-flow/internal/domain/entity"
)

// ClusterRepository 定义集群仓储接口
type ClusterRepository interface {
	// Create 创建集群
	Create(ctx context.Context, cluster *entity.Cluster) error

	// Update 更新集群
	Update(ctx context.Context, cluster *entity.Cluster) error

	// Delete 删除集群
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID 根据ID查找集群
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Cluster, error)

	// FindByEnvironmentID 根据环境ID查找集群
	FindByEnvironmentID(ctx context.Context, environmentID uuid.UUID) ([]*entity.Cluster, error)

	// ListWithFilters 根据条件列出集群
	ListWithFilters(ctx context.Context, environmentID *uuid.UUID, status *entity.ClusterStatus) ([]*entity.Cluster, error)

	// UpdateExecutorStatus 更新执行器状态
	UpdateExecutorStatus(ctx context.Context, id uuid.UUID, count int, lastHeartbeat time.Time) error
}
