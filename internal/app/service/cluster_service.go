package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/domain/entity"
	"mengri-flow/internal/domain/repository"

	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type clusterServiceImpl struct {
	clusterRepo     repository.ClusterRepository     `autowired:""`
	environmentRepo repository.EnvironmentRepository `autowired:""`
	flowRepo        repository.FlowRepository        `autowired:""`
	triggerRepo     repository.TriggerRepository     `autowired:""`
}

func (s *clusterServiceImpl) CreateCluster(ctx context.Context, req *dto.CreateClusterRequest) (*dto.ClusterResponse, error) {
	// 参数验证已经在 Gin 层面完成，这里进行业务验证
	slog.Info("Creating cluster", "name", req.Name, "environmentId", req.EnvironmentID)

	// 解析环境ID
	environmentID, err := uuid.Parse(req.EnvironmentID)
	if err != nil {
		slog.Warn("Invalid environment ID", "environmentId", req.EnvironmentID, "error", err)
		return nil, fmt.Errorf("invalid environment ID: %w", err)
	}

	// 检查环境是否存在
	env, err := s.environmentRepo.FindByID(ctx, environmentID)
	if err != nil {
		slog.Warn("Environment not found", "environmentId", environmentID, "error", err)
		return nil, fmt.Errorf("environment not found: %w", err)
	}
	if env == nil {
		return nil, fmt.Errorf("environment not found")
	}

	// 解析 etcd 端点字符串为数组
	etcdEndpoints := strings.Split(strings.TrimSpace(req.EtcdEndpoints), ",")
	for i := range etcdEndpoints {
		etcdEndpoints[i] = strings.TrimSpace(etcdEndpoints[i])
	}

	// 创建集群实体
	cluster, err := entity.NewCluster(
		req.Name,
		environmentID,
		etcdEndpoints,
		req.EtcdUsername,
		req.EtcdPassword,
		fmt.Sprintf("/mengri/%s/", environmentID.String()), // 自动生成前缀
		req.Description,
	)
	if err != nil {
		slog.Warn("Failed to create cluster entity", "error", err)
		return nil, fmt.Errorf("create cluster: %w", err)
	}

	// 保存到数据库
	err = s.clusterRepo.Create(ctx, cluster)
	if err != nil {
		slog.Error("Failed to save cluster to database", "error", err)
		return nil, fmt.Errorf("save cluster: %w", err)
	}

	slog.Info("Cluster created successfully", "clusterId", cluster.ID)
	return s.toClusterResponse(ctx, cluster), nil
}

func (s *clusterServiceImpl) ListClusters(ctx context.Context, environmentID string, page int, pageSize int) (*dto.ListClustersResponse, error) {
	slog.Info("Listing clusters", "environmentId", environmentID, "page", page, "pageSize", pageSize)

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// 这里需要实现带分页的查询
	// 由于 Repository 层没有提供分页查询接口，我们需要先获取所有数据然后手动分页
	// 或者修改 Repository 接口来支持分页
	// 暂时先实现简单的查询

	var envUUID *uuid.UUID
	if environmentID != "" {
		parsedEnvID, err := uuid.Parse(environmentID)
		if err != nil {
			slog.Warn("Invalid environment ID", "environmentId", environmentID, "error", err)
			return nil, fmt.Errorf("invalid environment ID: %w", err)
		}
		envUUID = &parsedEnvID
	}

	// 获取所有符合条件的集群
	clusters, err := s.clusterRepo.ListWithFilters(ctx, envUUID, nil)
	if err != nil {
		slog.Error("Failed to list clusters", "error", err)
		return nil, fmt.Errorf("list clusters: %w", err)
	}

	// 手动分页
	total := len(clusters)
	start := offset
	if start > total {
		start = total
	}
	end := offset + pageSize
	if end > total {
		end = total
	}

	// 转换响应
	responseList := make([]dto.ClusterResponse, 0, end-start)
	for i := start; i < end; i++ {
		responseList = append(responseList, *s.toClusterResponse(ctx, clusters[i]))
	}

	slog.Info("Clusters listed successfully", "total", total, "returned", len(responseList))
	return &dto.ListClustersResponse{
		Total:    int64(total),
		Page:     page,
		PageSize: pageSize,
		List:     responseList,
	}, nil
}

func (s *clusterServiceImpl) GetClusterDetail(ctx context.Context, id string) (*dto.ClusterDetailResponse, error) {
	slog.Info("Getting cluster detail", "clusterId", id)

	// 解析集群ID
	clusterID, err := uuid.Parse(id)
	if err != nil {
		slog.Warn("Invalid cluster ID", "clusterId", id, "error", err)
		return nil, fmt.Errorf("invalid cluster ID: %w", err)
	}

	// 获取集群
	cluster, err := s.clusterRepo.FindByID(ctx, clusterID)
	if err != nil {
		slog.Error("Failed to find cluster", "clusterId", clusterID, "error", err)
		return nil, fmt.Errorf("find cluster: %w", err)
	}
	if cluster == nil {
		slog.Warn("Cluster not found", "clusterId", clusterID)
		return nil, fmt.Errorf("cluster not found")
	}

	// 获取该集群的活跃流程数量（通过触发器查询）
	activeFlows := 0
	// 获取所有触发器（没有分页限制，所以使用大的 limit）
	triggers, _, err := s.triggerRepo.ListWithFilters(ctx, nil, nil, 0, 1000)
	if err == nil {
		for _, trigger := range triggers {
			// 这里简化处理：检查是否有触发器关联到该集群
			// 实际项目中可能需要更复杂的逻辑
			if trigger.Status == entity.TriggerStatusActive {
				activeFlows++
			}
		}
	}

	// 构建详细响应
	response := s.toClusterResponse(ctx, cluster)
	detailResponse := &dto.ClusterDetailResponse{
		ClusterResponse: *response,
		ExecutorCount:   cluster.ExecutorCount,
		ActiveFlows:     activeFlows,
	}

	slog.Info("Cluster detail retrieved successfully", "clusterId", clusterID)
	return detailResponse, nil
}

func (s *clusterServiceImpl) UpdateCluster(ctx context.Context, id string, req *dto.UpdateClusterRequest) (*dto.ClusterResponse, error) {
	slog.Info("Updating cluster", "clusterId", id)

	// 解析集群ID
	clusterID, err := uuid.Parse(id)
	if err != nil {
		slog.Warn("Invalid cluster ID", "clusterId", id, "error", err)
		return nil, fmt.Errorf("invalid cluster ID: %w", err)
	}

	// 获取集群
	cluster, err := s.clusterRepo.FindByID(ctx, clusterID)
	if err != nil {
		slog.Error("Failed to find cluster", "clusterId", clusterID, "error", err)
		return nil, fmt.Errorf("find cluster: %w", err)
	}
	if cluster == nil {
		slog.Warn("Cluster not found", "clusterId", clusterID)
		return nil, fmt.Errorf("cluster not found")
	}

	// 检查是否可以更新
	if cluster.Status == entity.ClusterStatusError {
		slog.Warn("Cannot update cluster with error status", "clusterId", clusterID, "status", cluster.Status)
		return nil, fmt.Errorf("cannot update cluster with error status")
	}

	// 构建更新数据
	updateName := req.Name
	if updateName == "" {
		updateName = cluster.Name
	}
	updateDescription := req.Description
	if updateDescription == "" {
		updateDescription = cluster.Description
	}

	// 更新集群实体
	err = cluster.Update(updateName, updateDescription)
	if err != nil {
		slog.Warn("Failed to update cluster entity", "error", err)
		return nil, fmt.Errorf("update cluster: %w", err)
	}

	// 保存到数据库
	err = s.clusterRepo.Update(ctx, cluster)
	if err != nil {
		slog.Error("Failed to save cluster update to database", "error", err)
		return nil, fmt.Errorf("save cluster: %w", err)
	}

	slog.Info("Cluster updated successfully", "clusterId", clusterID)
	return s.toClusterResponse(ctx, cluster), nil
}

func (s *clusterServiceImpl) DeleteCluster(ctx context.Context, id string) error {
	slog.Info("Deleting cluster", "clusterId", id)

	// 解析集群ID
	clusterID, err := uuid.Parse(id)
	if err != nil {
		slog.Warn("Invalid cluster ID", "clusterId", id, "error", err)
		return fmt.Errorf("invalid cluster ID: %w", err)
	}

	// 获取集群
	cluster, err := s.clusterRepo.FindByID(ctx, clusterID)
	if err != nil {
		slog.Error("Failed to find cluster", "clusterId", clusterID, "error", err)
		return fmt.Errorf("find cluster: %w", err)
	}
	if cluster == nil {
		slog.Warn("Cluster not found", "clusterId", clusterID)
		return fmt.Errorf("cluster not found")
	}

	// 检查是否可以删除
	if cluster.Status == entity.ClusterStatusActive {
		slog.Warn("Cannot delete active cluster", "clusterId", clusterID, "status", cluster.Status)
		return fmt.Errorf("cannot delete active cluster, please deactivate it first")
	}

	// 检查是否有活跃流程关联到该集群
	// 获取所有触发器（没有分页限制，所以使用大的 limit）
	triggers, _, err := s.triggerRepo.ListWithFilters(ctx, nil, nil, 0, 1000)
	if err != nil {
		slog.Error("Failed to check triggers", "error", err)
		return fmt.Errorf("check related triggers: %w", err)
	}

	hasActiveTriggers := false
	for _, trigger := range triggers {
		// 这里简化处理：检查所有活跃触发器
		// 实际项目中需要检查触发器是否属于当前集群
		if trigger.Status == entity.TriggerStatusActive {
			hasActiveTriggers = true
			break
		}
	}

	if hasActiveTriggers {
		slog.Warn("Cannot delete cluster with active triggers", "clusterId", clusterID)
		return fmt.Errorf("cannot delete cluster with active triggers")
	}

	// 删除集群
	err = s.clusterRepo.Delete(ctx, clusterID)
	if err != nil {
		slog.Error("Failed to delete cluster from database", "error", err)
		return fmt.Errorf("delete cluster: %w", err)
	}

	slog.Info("Cluster deleted successfully", "clusterId", clusterID)
	return nil
}

func (s *clusterServiceImpl) TestEtcdConnection(ctx context.Context, req *dto.TestEtcdConnectionRequest) (*dto.TestEtcdConnectionResponse, error) {
	// Create etcd client with provided endpoints
	cfg := clientv3.Config{
		Endpoints:   []string{req.Endpoints},
		Username:    req.Username,
		Password:    req.Password,
		DialTimeout: 5 * time.Second,
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		return &dto.TestEtcdConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create etcd client: %v", err),
		}, nil
	}
	defer client.Close()

	// Test connection by performing a simple operation
	_, err = client.Status(ctx, req.Endpoints)
	if err != nil {
		return &dto.TestEtcdConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to connect to etcd: %v", err),
		}, nil
	}

	return &dto.TestEtcdConnectionResponse{
		Success: true,
		Message: "Successfully connected to etcd",
	}, nil
}

// toClusterResponse 将集群实体转换为响应DTO
func (s *clusterServiceImpl) toClusterResponse(_ context.Context, cluster *entity.Cluster) *dto.ClusterResponse {
	// 将 etcd 端点数组转换为字符串
	etcdEndpoints := ""
	if len(cluster.EtcdEndpoints) > 0 {
		etcdEndpoints = strings.Join(cluster.EtcdEndpoints, ",")
	}

	// 根据执行器数量计算节点数
	nodeCount := max(cluster.ExecutorCount, 0)

	return &dto.ClusterResponse{
		ID:            cluster.ID.String(),
		Name:          cluster.Name,
		Description:   cluster.Description,
		EnvironmentID: cluster.EnvironmentID.String(),
		EtcdEndpoints: etcdEndpoints,
		EtcdUsername:  cluster.EtcdUsername,
		Status:        string(cluster.Status),
		NodeCount:     nodeCount,
		CreatedAt:     cluster.CreatedAt,
		UpdatedAt:     cluster.UpdatedAt,
	}
}
