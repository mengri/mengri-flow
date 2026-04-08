package clusterRepository

import (
	"context"
	"fmt"
	"time"

	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/repository"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

// ClusterRepositoryImpl 是 ClusterRepository 的 GORM实现
type ClusterRepositoryImpl struct {
	db *gorm.DB `autowired:""`
}

var _ repository.ClusterRepository = (*ClusterRepositoryImpl)(nil)

// Create 创建集群
func (r *ClusterRepositoryImpl) Create(ctx context.Context, cluster *entity.Cluster) error {
	model := toModel(cluster)
	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return fmt.Errorf("clusterRepository.Create: failed to create cluster: %w", result.Error)
	}
	cluster.ID = uuid.MustParse(model.ID)
	return nil
}

// Update 更新集群
func (r *ClusterRepositoryImpl) Update(ctx context.Context, cluster *entity.Cluster) error {
	model := toModel(cluster)
	result := r.db.WithContext(ctx).Model(&ClusterModel{}).Where("id = ?", model.ID).Updates(model)
	if result.Error != nil {
		return fmt.Errorf("clusterRepository.Update: failed to update cluster: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domainErr.ErrNotFound
	}
	return nil
}

// Delete 删除集群
func (r *ClusterRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&ClusterModel{}, "id = ?", id.String())
	if result.Error != nil {
		return fmt.Errorf("clusterRepository.Delete: failed to delete cluster: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domainErr.ErrNotFound
	}
	return nil
}

// FindByID 根据ID查找集群
func (r *ClusterRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.Cluster, error) {
	var model ClusterModel
	result := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErr.ErrNotFound
		}
		return nil, fmt.Errorf("clusterRepository.FindByID: failed to find cluster: %w", result.Error)
	}
	return toEntity(&model)
}

// FindByEnvironmentID 根据环境ID查找集群
func (r *ClusterRepositoryImpl) FindByEnvironmentID(ctx context.Context, environmentID uuid.UUID) ([]*entity.Cluster, error) {
	var models []ClusterModel
	result := r.db.WithContext(ctx).Where("environment_id = ?", environmentID.String()).
		Order("created_at DESC").Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("clusterRepository.FindByEnvironmentID: failed to find clusters: %w", result.Error)
	}

	clusters := make([]*entity.Cluster, len(models))
	for i, model := range models {
		cluster, err := toEntity(&model)
		if err != nil {
			return nil, err
		}
		clusters[i] = cluster
	}
	return clusters, nil
}

// ListWithFilters 根据条件列出集群
func (r *ClusterRepositoryImpl) ListWithFilters(ctx context.Context, environmentID *uuid.UUID, status *entity.ClusterStatus) ([]*entity.Cluster, error) {
	var models []ClusterModel
	query := r.db.WithContext(ctx).Model(&ClusterModel{})

	if environmentID != nil {
		query = query.Where("environment_id = ?", environmentID.String())
	}
	if status != nil {
		query = query.Where("status = ?", string(*status))
	}

	result := query.Order("created_at DESC").Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("clusterRepository.ListWithFilters: failed to list clusters: %w", result.Error)
	}

	clusters := make([]*entity.Cluster, len(models))
	for i, model := range models {
		cluster, err := toEntity(&model)
		if err != nil {
			return nil, err
		}
		clusters[i] = cluster
	}
	return clusters, nil
}

// UpdateExecutorStatus 更新执行器状态
func (r *ClusterRepositoryImpl) UpdateExecutorStatus(ctx context.Context, id uuid.UUID, count int, lastHeartbeat time.Time) error {
	result := r.db.WithContext(ctx).Model(&ClusterModel{}).
		Where("id = ?", id.String()).
		Updates(map[string]interface{}{
			"executor_count": count,
			"last_heartbeat": lastHeartbeat,
			"updated_at":     time.Now(),
		})
	if result.Error != nil {
		return fmt.Errorf("clusterRepository.UpdateExecutorStatus: failed to update executor status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domainErr.ErrNotFound
	}
	return nil
}
