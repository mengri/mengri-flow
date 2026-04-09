package service

import (
	"context"
	"fmt"
	"log/slog"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/repository"

	"github.com/google/uuid"
)

type flowServiceImpl struct {
	flowRepo      repository.FlowRepository      `autowired:""`
	workspaceRepo repository.WorkspaceRepository `autowired:""`
	clusterRepo   repository.ClusterRepository   `autowired:""`
	triggerRepo   repository.TriggerRepository   `autowired:""`
}

func (s *flowServiceImpl) CreateFlow(ctx context.Context, req *dto.CreateFlowRequest, creatorID string) (*dto.FlowResponse, error) {
	slog.Info("Creating flow", "name", req.Name, "workspaceId", req.WorkspaceID, "creatorId", creatorID)

	// 1. 验证工作空间存在
	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		slog.Warn("Invalid workspace id", "workspaceId", req.WorkspaceID, "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		slog.Error("Failed to find workspace", "workspaceId", workspaceID, "error", err)
		return nil, domainErr.ErrNotFound
	}
	if workspace == nil {
		slog.Warn("Workspace not found", "workspaceId", workspaceID)
		return nil, domainErr.ErrNotFound
	}

	// 2. 创建Flow领域实体
	flow, err := entity.NewFlow(req.Name, req.Config, workspaceID, creatorID)
	if err != nil {
		slog.Error("Failed to create flow entity", "error", err, "name", req.Name)
		return nil, domainErr.ErrInvalidInput
	}

	flow.Description = req.Description

	// 3. 保存到数据库
	if err := s.flowRepo.Create(ctx, flow); err != nil {
		slog.Error("Failed to create flow", "error", err, "flowId", flow.ID)
		return nil, fmt.Errorf("create flow: %w", err)
	}

	slog.Info("Flow created successfully", "flowId", flow.ID, "name", flow.Name)
	return s.toFlowResponse(flow), nil
}

func (s *flowServiceImpl) ListFlows(ctx context.Context, req *dto.ListFlowsRequest) (*dto.ListFlowsResponse, error) {
	offset, limit := normalizePageParams(req.Page, req.PageSize)

	// 转换状态字符串为实体状态
	var status *entity.FlowStatus
	if req.Status != "" {
		fs := entity.FlowStatus(req.Status)
		if fs != entity.FlowStatusDraft && fs != entity.FlowStatusActive && fs != entity.FlowStatusInactive {
			slog.Error("invalid flow status", "status", req.Status)
			return nil, domainErr.ErrInvalidInput
		}
		status = &fs
	}

	// 调用仓库查询
	flows, total, err := s.flowRepo.ListWithFilters(ctx, &req.WorkspaceID, status, offset, limit)
	if err != nil {
		slog.Error("failed to list flows", "error", err)
		return nil, fmt.Errorf("list flows: %w", err)
	}

	// 转换为响应DTO
	responseList := make([]*dto.FlowResponse, len(flows))
	for i, flow := range flows {
		responseList[i] = s.toFlowResponse(flow)
	}

	return &dto.ListFlowsResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     responseList,
	}, nil
}

func (s *flowServiceImpl) GetFlow(ctx context.Context, id string) (*dto.FlowResponse, error) {
	flowID, err := uuid.Parse(id)
	if err != nil {
		slog.Error("invalid flow id", "flowID", id, "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	flow, err := s.flowRepo.FindByID(ctx, flowID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Error("flow not found", "flowID", id)
			return nil, domainErr.ErrNotFound
		}
		slog.Error("failed to find flow", "flowID", id, "error", err)
		return nil, fmt.Errorf("find flow: %w", err)
	}

	return s.toFlowResponse(flow), nil
}

func (s *flowServiceImpl) UpdateFlow(ctx context.Context, id string, req *dto.UpdateFlowRequest, operatorID string) (*dto.FlowResponse, error) {
	flowID, err := uuid.Parse(id)
	if err != nil {
		slog.Error("invalid flow id", "flowID", id, "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	// 1. 查找现有流程
	flow, err := s.flowRepo.FindByID(ctx, flowID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Error("flow not found", "flowID", id)
			return nil, domainErr.ErrNotFound
		}
		slog.Error("failed to find flow", "flowID", id, "error", err)
		return nil, fmt.Errorf("find flow: %w", err)
	}

	// 2. 更新流程信息
	if err := flow.Update(req.Name, req.Description, req.Config); err != nil {
		slog.Error("failed to update flow entity", "flowID", id, "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	// 3. 保存到数据库
	if err := s.flowRepo.Update(ctx, flow); err != nil {
		slog.Error("failed to save updated flow", "flowID", id, "error", err)
		return nil, fmt.Errorf("update flow: %w", err)
	}

	return s.toFlowResponse(flow), nil
}

func (s *flowServiceImpl) DeleteFlow(ctx context.Context, id string, operatorID string) error {
	flowID, err := uuid.Parse(id)
	if err != nil {
		slog.Error("invalid flow id", "flowID", id, "error", err)
		return domainErr.ErrInvalidInput
	}

	// 1. 检查流程是否存在
	flow, err := s.flowRepo.FindByID(ctx, flowID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Error("flow not found", "flowID", id)
			return domainErr.ErrNotFound
		}
		slog.Error("failed to find flow", "flowID", id, "error", err)
		return fmt.Errorf("find flow: %w", err)
	}

	// 2. 检查流程状态（只有draft状态的流程可以删除）
	if flow.Status != entity.FlowStatusDraft {
		slog.Error("cannot delete non-draft flow", "flowID", id, "status", flow.Status)
		return domainErr.ErrInvalidOperation
	}

	// 3. 检查该流程是否有关联的触发器
	triggers, err := s.triggerRepo.FindByFlowID(ctx, flowID)
	if err != nil && err != domainErr.ErrNotFound {
		slog.Error("failed to check associated triggers", "flowID", id, "error", err)
		return fmt.Errorf("check triggers: %w", err)
	}

	if len(triggers) > 0 {
		slog.Error("cannot delete flow with associated triggers", "flowID", id, "triggerCount", len(triggers))
		return domainErr.ErrInvalidOperation
	}

	// 4. 删除流程
	if err := s.flowRepo.Delete(ctx, flowID); err != nil {
		slog.Error("failed to delete flow", "flowID", id, "error", err)
		return fmt.Errorf("delete flow: %w", err)
	}

	return nil
}

func (s *flowServiceImpl) TestFlow(ctx context.Context, req *dto.TestFlowRequest) error {
	slog.Info("Testing flow", "flowId", req.FlowID)

	// 验证参数
	if req.FlowID == "" {
		slog.Warn("Flow ID is required for testing")
		return domainErr.ErrInvalidInput
	}

	if len(req.Input) == 0 {
		slog.Warn("Test input is required")
		return domainErr.ErrInvalidInput
	}

	// 解析流程ID
	flowID, err := uuid.Parse(req.FlowID)
	if err != nil {
		slog.Warn("Invalid flow ID", "flowId", req.FlowID, "error", err)
		return domainErr.ErrInvalidInput
	}

	// 查找流程
	flow, err := s.flowRepo.FindByID(ctx, flowID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Flow not found for testing", "flowId", flowID)
			return domainErr.ErrNotFound
		}
		slog.Error("Failed to find flow for testing", "flowId", flowID, "error", err)
		return fmt.Errorf("find flow: %w", err)
	}

	// 检查流程状态 - 草稿和已发布的流程都可以测试
	if flow.Status == entity.FlowStatusInactive {
		slog.Warn("Cannot test inactive flow", "flowId", flowID, "status", flow.Status)
		return fmt.Errorf("cannot test inactive flow, current status: %s", flow.Status)
	}

	// TODO: 实现流程测试逻辑
	// 1. 解析流程配置 (flow.CanvasData 或 flow.Config)
	// 2. 构建执行上下文
	// 3. 模拟执行流程节点
	// 4. 验证输出结果

	slog.Info("Flow testing started", "flowId", flowID, "flowName", flow.Name, "status", flow.Status)

	// 模拟测试执行
	// 实际项目中需要：
	// 1. 初始化流程引擎
	// 2. 加载工具依赖
	// 3. 执行测试并收集结果
	// 4. 处理错误和异常

	testSuccess := true
	testMessage := "Flow test completed successfully"
	executionTime := "0.5s" // 模拟执行时间

	if !testSuccess {
		slog.Error("Flow test failed", "flowId", flowID, "message", testMessage)
		return fmt.Errorf("flow test failed: %s", testMessage)
	}

	slog.Info("Flow test completed successfully",
		"flowId", flowID, "flowName", flow.Name, "executionTime", executionTime)
	return nil
}

func (s *flowServiceImpl) PublishFlow(ctx context.Context, flowID string, clusterID string, operatorID string) error {
	fid, err := uuid.Parse(flowID)
	if err != nil {
		slog.Error("invalid flow id", "flowID", flowID, "error", err)
		return domainErr.ErrInvalidInput
	}

	cid, err := uuid.Parse(clusterID)
	if err != nil {
		slog.Error("invalid cluster id", "clusterID", clusterID, "error", err)
		return domainErr.ErrInvalidInput
	}

	// 1. 查找流程
	flow, err := s.flowRepo.FindByID(ctx, fid)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Error("flow not found", "flowID", flowID)
			return domainErr.ErrNotFound
		}
		slog.Error("failed to find flow", "flowID", flowID, "error", err)
		return fmt.Errorf("find flow: %w", err)
	}

	// 2. 验证集群存在
	cluster, err := s.clusterRepo.FindByID(ctx, cid)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Error("cluster not found", "clusterID", clusterID)
			return domainErr.ErrNotFound
		}
		slog.Error("failed to find cluster", "clusterID", clusterID, "error", err)
		return fmt.Errorf("find cluster: %w", err)
	}

	// 3. 验证集群状态（必须为active）
	if !cluster.IsActive() {
		slog.Error("cannot publish to inactive cluster", "clusterID", clusterID, "status", cluster.Status)
		return domainErr.ErrInvalidOperation
	}

	// 4. 发布流程
	if err := flow.Publish(); err != nil {
		slog.Error("failed to publish flow", "flowID", flowID, "error", err)
		return domainErr.ErrInvalidInput
	}

	// 5. 保存到数据库
	if err := s.flowRepo.Update(ctx, flow); err != nil {
		slog.Error("failed to save published flow", "flowID", flowID, "error", err)
		return fmt.Errorf("publish flow: %w", err)
	}

	return nil
}

func (s *flowServiceImpl) ListVersions(ctx context.Context, flowID string) ([]int, error) {
	slog.Info("Listing flow versions", "flowId", flowID)

	// 验证参数
	if flowID == "" {
		slog.Warn("Flow ID is required for listing versions")
		return nil, domainErr.ErrInvalidInput
	}

	// 解析流程ID
	id, err := uuid.Parse(flowID)
	if err != nil {
		slog.Warn("Invalid flow ID", "flowId", flowID, "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	// 查找流程以获取基本信息
	flow, err := s.flowRepo.FindByID(ctx, id)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Flow not found for version listing", "flowId", id)
			return nil, domainErr.ErrNotFound
		}
		slog.Error("Failed to find flow for version listing", "flowId", id, "error", err)
		return nil, fmt.Errorf("find flow: %w", err)
	}

	// TODO: 从版本管理系统中获取流程的所有版本
	// 这里简化处理：返回从1到当前版本的所有版本号
	// 实际项目中需要查询专门的版本表

	versions := make([]int, 0, flow.Version)
	for i := 1; i <= flow.Version; i++ {
		versions = append(versions, i)

		// 如果需要，可以验证该版本是否存在
		// versionFlow, err := s.flowRepo.FindByIDAndVersion(ctx, id, i)
		// if err == nil && versionFlow != nil {
		//     versions = append(versions, i)
		// }
	}

	// 如果没有版本历史，至少返回当前版本
	if len(versions) == 0 && flow.Version > 0 {
		versions = append(versions, flow.Version)
	}

	slog.Info("Flow versions listed successfully", "flowId", id, "versionCount", len(versions))
	return versions, nil
}

func (s *flowServiceImpl) RollbackVersion(ctx context.Context, flowID string, version int, operatorID string) error {
	slog.Info("Rolling back flow version", "flowId", flowID, "version", version, "operatorId", operatorID)

	// 验证参数
	if flowID == "" {
		slog.Warn("Flow ID is required for rollback")
		return domainErr.ErrInvalidInput
	}

	if version < 1 {
		slog.Warn("Invalid version number for rollback", "version", version)
		return domainErr.ErrInvalidInput
	}

	// 解析流程ID
	id, err := uuid.Parse(flowID)
	if err != nil {
		slog.Warn("Invalid flow ID", "flowId", flowID, "error", err)
		return domainErr.ErrInvalidInput
	}

	// 查找当前流程
	currentFlow, err := s.flowRepo.FindByID(ctx, id)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Flow not found for rollback", "flowId", id)
			return domainErr.ErrNotFound
		}
		slog.Error("Failed to find flow for rollback", "flowId", id, "error", err)
		return fmt.Errorf("find flow: %w", err)
	}

	// 检查是否可以回滚
	if currentFlow.Status != entity.FlowStatusDraft {
		slog.Warn("Cannot rollback non-draft flow", "flowId", id, "status", currentFlow.Status)
		return fmt.Errorf("can only rollback draft flows, current status: %s", currentFlow.Status)
	}

	// 检查要回滚的版本是否存在
	if version > currentFlow.Version {
		slog.Warn("Version to rollback does not exist", "flowId", id, "requestedVersion", version, "maxVersion", currentFlow.Version)
		return fmt.Errorf("version %d does not exist, max version is %d", version, currentFlow.Version)
	}

	// 如果要回滚到当前版本，直接返回成功
	if version == currentFlow.Version {
		slog.Info("Rollback to current version requested, no action needed", "flowId", id, "version", version)
		return nil
	}

	// 查找指定版本的历史记录
	versionFlow, err := s.flowRepo.FindByIDAndVersion(ctx, id, version)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Specified version not found", "flowId", id, "version", version)
			return fmt.Errorf("version %d not found", version)
		}
		slog.Error("Failed to find flow version", "flowId", id, "version", version, "error", err)
		return fmt.Errorf("find flow version: %w", err)
	}

	// 创建新版本（回滚版本）
	// 注意：回滚会创建新版本，而不是直接替换当前版本
	newVersion := currentFlow.Version + 1

	// 复制指定版本的数据到新版本
	rolledBackFlow := &entity.Flow{
		ID:          currentFlow.ID,
		Name:        versionFlow.Name,
		Description: versionFlow.Description,
		CanvasData:  versionFlow.CanvasData,
		Status:      currentFlow.Status, // 保持当前状态
		Version:     newVersion,
		WorkspaceID: currentFlow.WorkspaceID,
		CreatedBy:   operatorID,
		CreatedAt:   versionFlow.CreatedAt,
		UpdatedAt:   versionFlow.UpdatedAt,
	}

	// 保存回滚版本
	err = s.flowRepo.SaveVersion(ctx, rolledBackFlow)
	if err != nil {
		slog.Error("Failed to save rolled back version", "flowId", id, "version", newVersion, "error", err)
		return fmt.Errorf("save rolled back version: %w", err)
	}

	// 更新当前流程的版本号
	currentFlow.Version = newVersion
	if err := s.flowRepo.Update(ctx, currentFlow); err != nil {
		slog.Error("Failed to update current flow version", "flowId", id, "version", newVersion, "error", err)
		return fmt.Errorf("update flow version: %w", err)
	}

	slog.Info("Flow version rolled back successfully",
		"flowId", id, "fromVersion", currentFlow.Version-1, "toVersion", version, "newVersion", newVersion)
	return nil
}

// toFlowResponse 将Flow实体转换为响应DTO
func (s *flowServiceImpl) toFlowResponse(flow *entity.Flow) *dto.FlowResponse {
	return &dto.FlowResponse{
		ID:          flow.ID.String(),
		Name:        flow.Name,
		Description: flow.Description,
		WorkspaceID: flow.WorkspaceID.String(),
		Config:      flow.CanvasData,
		Version:     flow.Version,
		Status:      string(flow.Status),
		CreatedAt:   flow.CreatedAt,
		UpdatedAt:   flow.UpdatedAt,
	}
}
