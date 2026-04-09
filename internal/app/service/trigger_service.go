package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/repository"
)

type triggerServiceImpl struct {
	triggerRepo   repository.TriggerRepository   `autowired:""`
	flowRepo      repository.FlowRepository      `autowired:""`
	clusterRepo   repository.ClusterRepository   `autowired:""`
	workspaceRepo repository.WorkspaceRepository `autowired:""`
}

func (s *triggerServiceImpl) CreateTrigger(ctx context.Context, req *dto.CreateTriggerRequest, creatorID string) (*dto.TriggerResponse, error) {
	// 1. 验证参数
	if req.Name == "" {
		return nil, domainErr.ErrInvalidInput
	}

	// 2. 验证FlowID格式并尝试转换为UUID
	flowID, err := uuid.Parse(req.FlowID)
	if err != nil {
		return nil, domainErr.ErrInvalidInput
	}

	// 3. 验证Trigger类型
	triggerType := entity.TriggerType(req.Type)
	if err := entity.ValidateTriggerType(triggerType); err != nil {
		return nil, fmt.Errorf("invalid trigger type: %w", err)
	}

	// 4. 解析可选字段
	var clusterID *uuid.UUID
	if req.ClusterID != "" {
		cid, err := uuid.Parse(req.ClusterID)
		if err != nil {
			return nil, domainErr.ErrInvalidInput
		}
		clusterID = &cid
	}

	var workspaceID *uuid.UUID
	if req.WorkspaceID != "" {
		wid, err := uuid.Parse(req.WorkspaceID)
		if err != nil {
			return nil, domainErr.ErrInvalidInput
		}
		workspaceID = &wid
	}

	// 5. 构建错误处理配置
	var errorHandling *entity.ErrorHandling
	if req.ErrorHandling != nil {
		errorHandling = &entity.ErrorHandling{
			Strategy:           req.ErrorHandling.Strategy,
			RetryOnFailure:     req.ErrorHandling.RetryOnFailure,
			CustomErrorFormat: req.ErrorHandling.CustomErrorFormat,
		}
	}

	// 6. 根据memories中的提示处理不同触发器类型的额外验证
	// RESTful触发器需要特定的配置验证，这里简化处理
	switch triggerType {
	case entity.TriggerTypeRESTful:
		if _, hasPath := req.Config["path"]; !hasPath {
			return nil, fmt.Errorf("RESTful trigger requires 'path' configuration")
		}
	case entity.TriggerTypeTimer:
		if _, hasSchedule := req.Config["schedule"]; !hasSchedule {
			return nil, fmt.Errorf("Timer trigger requires 'schedule' configuration")
		}
	case entity.TriggerTypeMQ:
		if _, hasTopic := req.Config["topic"]; !hasTopic {
			return nil, fmt.Errorf("MQ trigger requires 'topic' configuration")
		}
	}

	// 7. 创建触发器领域实体
	trigger, err := entity.NewTriggerWithOptions(
		req.Name,
		triggerType,
		flowID,
		req.Config,
		clusterID,
		workspaceID,
		errorHandling,
	)
	if err != nil {
		return nil, fmt.Errorf("trigger validation failed: %w", err)
	}

	// 设置 FlowVersion 和映射字段
	if req.FlowVersion > 0 {
		trigger.FlowVersion = req.FlowVersion
	}
	if req.InputMapping != nil {
		trigger.InputMapping = req.InputMapping
	}
	if req.OutputMapping != nil {
		trigger.OutputMapping = req.OutputMapping
	}

	// 8. 保存到数据库
	if err := s.triggerRepo.Create(ctx, trigger); err != nil {
		slog.Error("failed to create trigger", "trigger", req.Name, "error", err)
		return nil, fmt.Errorf("failed to create trigger: %w", err)
	}

	// 9. 转换为DTO返回
	return toTriggerResponse(trigger), nil
}

func (s *triggerServiceImpl) ListTriggers(ctx context.Context, req *dto.ListTriggersRequest) (*dto.ListTriggersResponse, error) {
	slog.Info("Listing triggers", "flowId", req.FlowID, "page", req.Page, "pageSize", req.PageSize)

	// 验证分页参数
	page := req.Page
	pageSize := req.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// 构建查询过滤器
	var flowIDFilter *string
	if req.FlowID != "" {
		flowIDFilter = &req.FlowID
	}
	var statusFilter *entity.TriggerStatus
	if req.Status != "" {
		status := entity.TriggerStatus(req.Status)
		// 手动验证触发器状态，因为entity包中没有公开的验证函数
		switch status {
		case entity.TriggerStatusActive, entity.TriggerStatusInactive, entity.TriggerStatusError:
			// 有效状态
			statusFilter = &status
		default:
			slog.Warn("Invalid trigger status filter", "status", req.Status)
			return nil, fmt.Errorf("invalid trigger status: %s", req.Status)
		}
	}

	// 调用仓库查询
	triggers, total, err := s.triggerRepo.ListWithFilters(ctx, flowIDFilter, statusFilter, offset, pageSize)
	if err != nil {
		slog.Error("Failed to list triggers", "error", err)
		return nil, fmt.Errorf("list triggers: %w", err)
	}

	// 转换为响应DTO
	responseList := make([]dto.TriggerResponse, 0, len(triggers))
	for _, trigger := range triggers {
		responseList = append(responseList, *toTriggerResponse(trigger))
	}

	slog.Info("Triggers listed successfully", "total", total, "returned", len(responseList))
	return &dto.ListTriggersResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		List:     responseList,
	}, nil
}

func (s *triggerServiceImpl) GetTrigger(ctx context.Context, id string) (*dto.TriggerResponse, error) {
	slog.Info("Getting trigger detail", "triggerId", id)

	// 验证参数
	if id == "" {
		return nil, domainErr.ErrInvalidInput
	}

	triggerID, err := uuid.Parse(id)
	if err != nil {
		slog.Warn("Invalid trigger ID", "triggerId", id, "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	// 查找触发器
	trigger, err := s.triggerRepo.FindByID(ctx, triggerID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Trigger not found", "triggerId", triggerID)
			return nil, domainErr.ErrNotFound
		}
		slog.Error("Failed to find trigger", "triggerId", triggerID, "error", err)
		return nil, fmt.Errorf("find trigger: %w", err)
	}

	slog.Info("Trigger retrieved successfully", "triggerId", triggerID)
	return toTriggerResponse(trigger), nil
}

func (s *triggerServiceImpl) UpdateTrigger(ctx context.Context, id string, req *dto.UpdateTriggerRequest, operatorID string) (*dto.TriggerResponse, error) {
	slog.Info("Updating trigger", "triggerId", id, "operatorId", operatorID)

	// 验证参数
	if id == "" {
		return nil, domainErr.ErrInvalidInput
	}

	triggerID, err := uuid.Parse(id)
	if err != nil {
		slog.Warn("Invalid trigger ID", "triggerId", id, "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	// 查找触发器
	trigger, err := s.triggerRepo.FindByID(ctx, triggerID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Trigger not found", "triggerId", triggerID)
			return nil, domainErr.ErrNotFound
		}
		slog.Error("Failed to find trigger", "triggerId", triggerID, "error", err)
		return nil, fmt.Errorf("find trigger: %w", err)
	}

	// 检查是否可以更新 - 活跃状态的触发器可能有限制
	if trigger.Status == entity.TriggerStatusActive {
		slog.Warn("Attempting to update active trigger", "triggerId", triggerID, "status", trigger.Status)
		// 可能允许更新，但需要注意可能影响正在运行的流程
	}

	// 构建更新数据
	updatedName := req.Name
	if updatedName == "" {
		updatedName = trigger.Name
	}
	updatedConfig := req.Config
	if updatedConfig == nil {
		updatedConfig = trigger.Config
	}

	// TODO: 检查配置的有效性，针对不同类型的触发器
	// 根据memories中的提示，不同类型的触发器需要不同的配置验证

	// 更新触发器
	err = trigger.Update(updatedName, updatedConfig)
	if err != nil {
		slog.Warn("Failed to update trigger entity", "triggerId", triggerID, "error", err)
		return nil, fmt.Errorf("update trigger: %w", err)
	}

	// 保存到数据库
	err = s.triggerRepo.Update(ctx, trigger)
	if err != nil {
		slog.Error("Failed to save trigger update to database", "triggerId", triggerID, "error", err)
		return nil, fmt.Errorf("save trigger: %w", err)
	}

	slog.Info("Trigger updated successfully", "triggerId", triggerID)
	return toTriggerResponse(trigger), nil
}

func (s *triggerServiceImpl) DeleteTrigger(ctx context.Context, id string, operatorID string) error {
	slog.Info("Deleting trigger", "triggerId", id, "operatorId", operatorID)

	// 1. 验证参数
	if id == "" {
		return domainErr.ErrInvalidInput
	}

	triggerID, err := uuid.Parse(id)
	if err != nil {
		slog.Warn("Invalid trigger ID", "triggerId", id, "error", err)
		return domainErr.ErrInvalidInput
	}

	// 2. 检查触发器是否存在
	trigger, err := s.triggerRepo.FindByID(ctx, triggerID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Trigger not found", "triggerId", triggerID)
			return domainErr.ErrNotFound
		}
		slog.Error("Failed to find trigger", "triggerId", triggerID, "error", err)
		return fmt.Errorf("find trigger: %w", err)
	}

	// 3. 验证权限（简化处理，实际需要验证操作者权限）
	// 这里简化权限检查，实际项目中应该有更复杂的权限验证逻辑
	slog.Info("Permission check for trigger deletion - simplified implementation",
		"triggerId", triggerID, "operatorId", operatorID)

	// 4. 检查触发器状态 - 不允许删除活跃状态的触发器
	if trigger.Status == entity.TriggerStatusActive {
		slog.Warn("Cannot delete active trigger", "triggerId", triggerID, "status", trigger.Status)
		return fmt.Errorf("cannot delete active trigger, disable it first")
	}

	// 5. 删除触发器
	if err := s.triggerRepo.Delete(ctx, triggerID); err != nil {
		slog.Error("Failed to delete trigger", "triggerId", triggerID, "error", err)
		return fmt.Errorf("delete trigger: %w", err)
	}

	slog.Info("Trigger deleted successfully", "triggerId", triggerID, "operatorId", operatorID)
	return nil
}

func (s *triggerServiceImpl) EnableTrigger(ctx context.Context, id string, operatorID string) error {
	slog.Info("Enabling trigger", "triggerId", id, "operatorId", operatorID)

	// 验证参数
	if id == "" {
		return domainErr.ErrInvalidInput
	}

	triggerID, err := uuid.Parse(id)
	if err != nil {
		slog.Warn("Invalid trigger ID", "triggerId", id, "error", err)
		return domainErr.ErrInvalidInput
	}

	// 查找触发器
	trigger, err := s.triggerRepo.FindByID(ctx, triggerID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Trigger not found", "triggerId", triggerID)
			return domainErr.ErrNotFound
		}
		slog.Error("Failed to find trigger", "triggerId", triggerID, "error", err)
		return fmt.Errorf("find trigger: %w", err)
	}

	// 检查当前状态
	if trigger.Status == entity.TriggerStatusActive {
		slog.Info("Trigger already enabled", "triggerId", triggerID)
		return nil // 已经是启用状态，直接返回成功
	}

	// 检查流程状态 - 确保关联的流程是可用的
	flow, err := s.flowRepo.FindByID(ctx, trigger.FlowID)
	if err != nil {
		slog.Error("Failed to find associated flow", "flowId", trigger.FlowID, "error", err)
		return fmt.Errorf("find flow: %w", err)
	}
	if flow == nil {
		slog.Warn("Associated flow not found", "flowId", trigger.FlowID)
		return fmt.Errorf("associated flow not found")
	}

	// 检查流程是否是已发布状态
	// TODO: 需要根据项目实际需求实现流程状态检查
	// flow.CheckStatus() 或其他状态检查方法

	// 根据memories中的提示，不同类型的触发器启用逻辑不同
	switch trigger.Type {
	case entity.TriggerTypeRESTful:
		// RESTful触发器需要启动HTTP监听器
		slog.Info("Enabling RESTful trigger - would start HTTP listener", "triggerId", triggerID)
	case entity.TriggerTypeTimer:
		// Timer触发器需要启动定时任务
		// 根据memories：需要通过etcd分布式锁保证一个集群一次触发只执行一次
		slog.Info("Enabling Timer trigger - would schedule timer with etcd lock", "triggerId", triggerID)
	case entity.TriggerTypeMQ:
		// MQ触发器需要启动消息监听器
		// 根据memories：每收到一条消息执行一次流程，直接在当前线程执行
		slog.Info("Enabling MQ trigger - would start MQ consumer", "triggerId", triggerID)
	}

	// 更新触发器状态为激活
	err = trigger.UpdateStatus(entity.TriggerStatusActive)
	if err != nil {
		slog.Warn("Failed to update trigger status", "triggerId", triggerID, "error", err)
		return fmt.Errorf("update trigger status: %w", err)
	}

	// 保存到数据库
	err = s.triggerRepo.Update(ctx, trigger)
	if err != nil {
		slog.Error("Failed to save trigger status to database", "triggerId", triggerID, "error", err)
		return fmt.Errorf("save trigger: %w", err)
	}

	slog.Info("Trigger enabled successfully", "triggerId", triggerID)
	return nil
}

func (s *triggerServiceImpl) DisableTrigger(ctx context.Context, id string, operatorID string) error {
	slog.Info("Disabling trigger", "triggerId", id, "operatorId", operatorID)

	// 验证参数
	if id == "" {
		return domainErr.ErrInvalidInput
	}

	triggerID, err := uuid.Parse(id)
	if err != nil {
		slog.Warn("Invalid trigger ID", "triggerId", id, "error", err)
		return domainErr.ErrInvalidInput
	}

	// 查找触发器
	trigger, err := s.triggerRepo.FindByID(ctx, triggerID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Trigger not found", "triggerId", triggerID)
			return domainErr.ErrNotFound
		}
		slog.Error("Failed to find trigger", "triggerId", triggerID, "error", err)
		return fmt.Errorf("find trigger: %w", err)
	}

	// 检查当前状态
	if trigger.Status != entity.TriggerStatusActive {
		slog.Info("Trigger already disabled or in error state", "triggerId", triggerID, "status", trigger.Status)
		return nil // 已经是非激活状态，直接返回成功
	}

	// 根据memories中的提示，不同类型的触发器禁用逻辑不同
	switch trigger.Type {
	case entity.TriggerTypeRESTful:
		// RESTful触发器需要停止HTTP监听器
		slog.Info("Disabling RESTful trigger - would stop HTTP listener", "triggerId", triggerID)
	case entity.TriggerTypeTimer:
		// Timer触发器需要停止定时任务
		slog.Info("Disabling Timer trigger - would cancel timer", "triggerId", triggerID)
	case entity.TriggerTypeMQ:
		// MQ触发器需要停止消息监听器
		slog.Info("Disabling MQ trigger - would stop MQ consumer", "triggerId", triggerID)
	}

	// 更新触发器状态为未激活
	err = trigger.UpdateStatus(entity.TriggerStatusInactive)
	if err != nil {
		slog.Warn("Failed to update trigger status", "triggerId", triggerID, "error", err)
		return fmt.Errorf("update trigger status: %w", err)
	}

	// 保存到数据库
	err = s.triggerRepo.Update(ctx, trigger)
	if err != nil {
		slog.Error("Failed to save trigger status to database", "triggerId", triggerID, "error", err)
		return fmt.Errorf("save trigger: %w", err)
	}

	slog.Info("Trigger disabled successfully", "triggerId", triggerID)
	return nil
}

func (s *triggerServiceImpl) PublishToCluster(ctx context.Context, triggerID string, clusterID string, operatorID string) error {
	slog.Info("Publishing trigger to cluster", "triggerId", triggerID, "clusterId", clusterID, "operatorId", operatorID)

	// 验证参数
	if triggerID == "" || clusterID == "" {
		return domainErr.ErrInvalidInput
	}

	parsedTriggerID, err := uuid.Parse(triggerID)
	if err != nil {
		slog.Warn("Invalid trigger ID", "triggerId", triggerID, "error", err)
		return domainErr.ErrInvalidInput
	}

	parsedClusterID, err := uuid.Parse(clusterID)
	if err != nil {
		slog.Warn("Invalid cluster ID", "clusterId", clusterID, "error", err)
		return domainErr.ErrInvalidInput
	}

	// 查找触发器
	trigger, err := s.triggerRepo.FindByID(ctx, parsedTriggerID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Trigger not found", "triggerId", parsedTriggerID)
			return domainErr.ErrNotFound
		}
		slog.Error("Failed to find trigger", "triggerId", parsedTriggerID, "error", err)
		return fmt.Errorf("find trigger: %w", err)
	}

	// 查找集群
	cluster, err := s.clusterRepo.FindByID(ctx, parsedClusterID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Cluster not found", "clusterId", parsedClusterID)
			return domainErr.ErrNotFound
		}
		slog.Error("Failed to find cluster", "clusterId", parsedClusterID, "error", err)
		return fmt.Errorf("find cluster: %w", err)
	}

	// 检查集群状态
	if cluster.Status != entity.ClusterStatusActive {
		slog.Warn("Cluster is not active", "clusterId", parsedClusterID, "status", cluster.Status)
		return fmt.Errorf("cluster is not active, current status: %s", cluster.Status)
	}

	// 检查触发器状态
	if trigger.Status != entity.TriggerStatusActive {
		slog.Warn("Trigger is not active", "triggerId", parsedTriggerID, "status", trigger.Status)
		return fmt.Errorf("trigger must be active before publishing, current status: %s", trigger.Status)
	}

	// 查找关联的流程
	flow, err := s.flowRepo.FindByID(ctx, trigger.FlowID)
	if err != nil {
		slog.Error("Failed to find associated flow", "flowId", trigger.FlowID, "error", err)
		return fmt.Errorf("find flow: %w", err)
	}
	if flow == nil {
		slog.Warn("Associated flow not found", "flowId", trigger.FlowID)
		return fmt.Errorf("associated flow not found")
	}

	// TODO: 根据memories中的提示，不同类型的触发器发布到集群的逻辑不同
	switch trigger.Type {
	case entity.TriggerTypeRESTful:
		// RESTful触发器发布：执行器在指定端口创建目标path接口
		// 需要提供同步和异步两种接口范式
		slog.Info("Publishing RESTful trigger - would deploy HTTP endpoint to cluster",
			"triggerId", parsedTriggerID, "clusterId", parsedClusterID)
	case entity.TriggerTypeTimer:
		// Timer触发器发布：需要通过etcd分布式锁保证一个集群一次触发只执行一次
		slog.Info("Publishing Timer trigger - would deploy timer with etcd lock to cluster",
			"triggerId", parsedTriggerID, "clusterId", parsedClusterID)
	case entity.TriggerTypeMQ:
		// MQ触发器发布：监听MQ指定地址、topic和消费策略
		// 每收到一条消息执行一次流程，不需要异步执行，直接在当前线程执行
		slog.Info("Publishing MQ trigger - would deploy MQ consumer to cluster",
			"triggerId", parsedTriggerID, "clusterId", parsedClusterID)
	}

	// TODO: 实现实际的发布逻辑
	// 1. 生成触发器配置
	// 2. 将配置推送到集群的etcd
	// 3. 等待集群执行器确认接收
	// 4. 更新触发器元数据（如上次发布时间、目标集群等）

	slog.Warn("Trigger publish to cluster not fully implemented - simulated success",
		"triggerId", parsedTriggerID, "clusterId", parsedClusterID)

	// 由于完整的发布逻辑需要与集群执行器交互，这里返回成功模拟
	// 实际项目中需要实现完整的发布流程
	return fmt.Errorf("trigger publish to cluster not fully implemented")
}

// --- 私有方法 ---

// toTriggerResponse 将领域实体转换为DTO响应
func toTriggerResponse(trigger *entity.Trigger) *dto.TriggerResponse {
	var errorHandling *dto.ErrorHandlingResponse
	if trigger.ErrorHandling.Strategy != "" {
		errorHandling = &dto.ErrorHandlingResponse{
			Strategy:           trigger.ErrorHandling.Strategy,
			CustomErrorFormat: trigger.ErrorHandling.CustomErrorFormat,
			RetryOnFailure:     trigger.ErrorHandling.RetryOnFailure,
		}
	}

	return &dto.TriggerResponse{
		ID:            trigger.ID.String(),
		Name:          trigger.Name,
		Type:          string(trigger.Type),
		Config:        trigger.Config,
		FlowID:        trigger.FlowID.String(),
		FlowVersion:   trigger.FlowVersion,
		ClusterID:     trigger.ClusterID.String(),
		InputMapping:  trigger.InputMapping,
		OutputMapping: trigger.OutputMapping,
		ErrorHandling: errorHandling,
		WorkspaceID:   trigger.WorkspaceID.String(),
		Status:        string(trigger.Status),
		Description:   "", // Trigger实体没有Description字段
		CreatedAt:     trigger.CreatedAt,
		UpdatedAt:     trigger.UpdatedAt,
	}
}
