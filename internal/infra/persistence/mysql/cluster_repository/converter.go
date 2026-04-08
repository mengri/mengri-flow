package clusterRepository

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"mengri-flow/internal/domain/entity"
)

// toModel 将领域实体转换为GORM模型
func toModel(cluster *entity.Cluster) *ClusterModel {
	endpointsJSON, _ := json.Marshal(cluster.EtcdEndpoints)

	return &ClusterModel{
		ID:             cluster.ID.String(),
		Name:           cluster.Name,
		EnvironmentID:  cluster.EnvironmentID.String(),
		EtcdEndpoints:  string(endpointsJSON),
		EtcdUsername:   cluster.EtcdUsername,
		EtcdPassword:   cluster.EtcdPassword,
		EtcdPrefix:     cluster.EtcdPrefix,
		Description:    cluster.Description,
		Status:         string(cluster.Status),
		ExecutorCount:  cluster.ExecutorCount,
		LastHeartbeat:  cluster.LastHeartbeat,
		CreatedAt:      cluster.CreatedAt,
		UpdatedAt:      cluster.UpdatedAt,
	}
}

// toEntity 将GORM模型转换为领域实体
func toEntity(model *ClusterModel) (*entity.Cluster, error) {
	var endpoints []string
	if err := json.Unmarshal([]byte(model.EtcdEndpoints), &endpoints); err != nil {
		return nil, fmt.Errorf("failed to unmarshal etcd endpoints: %w", err)
	}

	return &entity.Cluster{
		ID:             uuid.MustParse(model.ID),
		Name:           model.Name,
		EnvironmentID:  uuid.MustParse(model.EnvironmentID),
		EtcdEndpoints:  endpoints,
		EtcdUsername:   model.EtcdUsername,
		EtcdPassword:   model.EtcdPassword,
		EtcdPrefix:     model.EtcdPrefix,
		Description:    model.Description,
		Status:         entity.ClusterStatus(model.Status),
		ExecutorCount:  model.ExecutorCount,
		LastHeartbeat:  model.LastHeartbeat,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}, nil
}
