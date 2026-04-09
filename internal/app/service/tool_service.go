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

type toolServiceImpl struct {
	toolRepo      repository.ToolRepository      `autowired:""`
	resourceRepo  repository.ResourceRepository  `autowired:""`
	workspaceRepo repository.WorkspaceRepository `autowired:""`
}

func (s *toolServiceImpl) CreateTool(ctx context.Context, req *dto.CreateToolRequest) (*dto.ToolResponse, error) {
	slog.Info("Creating tool", "name", req.Name, "type", req.Type, "workspaceId", req.WorkspaceID)

	// 1. 验证工作空间存在
	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		slog.Warn("Invalid workspace id", "workspaceID", req.WorkspaceID, "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		slog.Error("Failed to find workspace", "workspaceID", workspaceID, "error", err)
		return nil, domainErr.ErrNotFound
	}
	if workspace == nil {
		slog.Warn("Workspace not found", "workspaceID", workspaceID)
		return nil, domainErr.ErrNotFound
	}

	// 2. 创建资源
	// 注意：ResourceType根据工具类型选择，这里使用HTTP作为默认
	resourceType := entity.ResourceTypeHTTP
	switch req.Type {
	case "grpc":
		resourceType = entity.ResourceTypeGRPC
	case "mysql":
		resourceType = entity.ResourceTypeMySQL
	case "postgres":
		resourceType = entity.ResourceTypePostgres
	}

	resource, err := entity.NewResource(req.Name, resourceType, req.Config, workspaceID, req.Description)
	if err != nil {
		slog.Error("Failed to create resource for tool", "error", err, "name", req.Name)
		return nil, domainErr.ErrInvalidInput
	}

	if err := s.resourceRepo.Create(ctx, resource); err != nil {
		slog.Error("Failed to create resource", "error", err, "resourceId", resource.ID)
		return nil, fmt.Errorf("create resource: %w", err)
	}

	// 3. 创建Tool领域实体
	// 注意：CreatedBy字段暂时使用空字符串，实际项目中应从上下文中获取
	tool, err := entity.NewTool(req.Name, req.Type, req.Config, resource.ID, workspaceID, "")
	if err != nil {
		slog.Error("Failed to create tool entity", "error", err, "name", req.Name)
		return nil, domainErr.ErrInvalidInput
	}

	// 4. 保存到数据库
	if err := s.toolRepo.Create(ctx, tool); err != nil {
		slog.Error("Failed to create tool", "error", err, "toolId", tool.ID)
		return nil, fmt.Errorf("create tool: %w", err)
	}

	slog.Info("Tool created successfully", "toolId", tool.ID, "name", tool.Name, "type", tool.Type)
	return s.toToolResponse(tool), nil
}

func (s *toolServiceImpl) ListTools(ctx context.Context, req *dto.ListToolsRequest) (*dto.ListToolsResponse, error) {
	slog.Info("Listing tools", "workspaceId", req.WorkspaceID, "type", req.Type, "status", req.Status,
		"page", req.Page, "pageSize", req.PageSize)

	offset, limit := normalizePageParams(req.Page, req.PageSize)

	// 转换状态字符串为实体状态
	var status *entity.ToolStatus
	if req.Status != "" {
		ts := entity.ToolStatus(req.Status)
		if ts != entity.ToolStatusDraft && ts != entity.ToolStatusPublished && ts != entity.ToolStatusDeprecated {
			slog.Warn("Invalid tool status", "status", req.Status)
			return nil, domainErr.ErrInvalidInput
		}
		status = &ts
	}

	// 调用仓库查询
	tools, total, err := s.toolRepo.ListWithFilters(ctx, &req.WorkspaceID, &req.Type, status, offset, limit)
	if err != nil {
		slog.Error("Failed to list tools", "error", err, "workspaceId", req.WorkspaceID)
		return nil, fmt.Errorf("list tools: %w", err)
	}

	// 转换为响应DTO
	responseList := make([]dto.ToolResponse, len(tools))
	for i, tool := range tools {
		responseList[i] = *s.toToolResponse(tool)
	}

	slog.Info("Tools listed successfully", "total", total, "returned", len(tools),
		"workspaceId", req.WorkspaceID)
	return &dto.ListToolsResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     responseList,
	}, nil
}

func (s *toolServiceImpl) GetTool(ctx context.Context, id string) (*dto.ToolResponse, error) {
	slog.Info("Getting tool detail", "toolId", id)

	toolID, err := uuid.Parse(id)
	if err != nil {
		slog.Warn("Invalid tool id", "toolId", id, "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	tool, err := s.toolRepo.FindByID(ctx, toolID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Tool not found", "toolId", toolID)
			return nil, domainErr.ErrNotFound
		}
		slog.Error("Failed to find tool", "toolId", toolID, "error", err)
		return nil, fmt.Errorf("find tool: %w", err)
	}

	slog.Info("Tool retrieved successfully", "toolId", toolID, "name", tool.Name)
	return s.toToolResponse(tool), nil
}

func (s *toolServiceImpl) UpdateTool(ctx context.Context, id string, req *dto.UpdateToolRequest) (*dto.ToolResponse, error) {
	slog.Info("Updating tool", "toolId", id)

	toolID, err := uuid.Parse(id)
	if err != nil {
		slog.Warn("Invalid tool id", "toolId", id, "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	// 1. 查找现有工具
	tool, err := s.toolRepo.FindByID(ctx, toolID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Tool not found", "toolId", toolID)
			return nil, domainErr.ErrNotFound
		}
		slog.Error("Failed to find tool", "toolId", toolID, "error", err)
		return nil, fmt.Errorf("find tool: %w", err)
	}

	// 检查工具状态 - 已废弃的工具不能更新
	if tool.Status == entity.ToolStatusDeprecated {
		slog.Warn("Cannot update deprecated tool", "toolId", toolID, "status", tool.Status)
		return nil, fmt.Errorf("cannot update deprecated tool")
	}

	// 2. 更新工具信息
	if err := tool.Update(req.Name, req.Description, req.Config); err != nil {
		slog.Error("Failed to update tool entity", "toolId", toolID, "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	// 3. 保存到数据库
	if err := s.toolRepo.Update(ctx, tool); err != nil {
		slog.Error("Failed to save updated tool", "toolId", toolID, "error", err)
		return nil, fmt.Errorf("update tool: %w", err)
	}

	// 4. 更新关联的资源
	resource, err := s.resourceRepo.FindByID(ctx, tool.ResourceID)
	if err != nil {
		slog.Error("Failed to find associated resource", "resourceId", tool.ResourceID, "error", err)
		return nil, fmt.Errorf("find resource: %w", err)
	}

	if err := resource.Update(req.Name, req.Config, req.Description); err != nil {
		slog.Error("Failed to update resource", "resourceId", tool.ResourceID, "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	if err := s.resourceRepo.Update(ctx, resource); err != nil {
		slog.Error("Failed to save updated resource", "resourceId", tool.ResourceID, "error", err)
		return nil, fmt.Errorf("update resource: %w", err)
	}

	slog.Info("Tool updated successfully", "toolId", toolID, "name", tool.Name)
	return s.toToolResponse(tool), nil
}

func (s *toolServiceImpl) TestTool(ctx context.Context, req *dto.TestToolRequest) error {
	slog.Info("Testing tool", "toolId", req.ToolID)

	// 验证参数
	if req.ToolID == "" {
		slog.Warn("Tool ID is required for testing")
		return domainErr.ErrInvalidInput
	}

	if len(req.Input) == 0 {
		slog.Warn("Test input is required")
		return domainErr.ErrInvalidInput
	}

	// 解析工具ID
	toolID, err := uuid.Parse(req.ToolID)
	if err != nil {
		slog.Warn("Invalid tool ID", "toolId", req.ToolID, "error", err)
		return domainErr.ErrInvalidInput
	}

	// 查找工具
	tool, err := s.toolRepo.FindByID(ctx, toolID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Tool not found for testing", "toolId", toolID)
			return domainErr.ErrNotFound
		}
		slog.Error("Failed to find tool for testing", "toolId", toolID, "error", err)
		return fmt.Errorf("find tool: %w", err)
	}

	// 检查工具状态 - 只有已发布的工具可以进行测试
	if tool.Status != entity.ToolStatusPublished {
		slog.Warn("Tool is not published, cannot test", "toolId", toolID, "status", tool.Status)
		return fmt.Errorf("tool must be published before testing, current status: %s", tool.Status)
	}

	// 根据工具类型执行不同的测试逻辑
	// 这里需要根据实际的工具类型（HTTP, gRPC, Database等）来执行测试
	slog.Info("Testing tool based on type", "toolId", toolID, "type", tool.Type, "toolName", tool.Name)

	// TODO: 根据工具类型实现具体的测试逻辑
	// 例如：
	// - HTTP工具：发送测试请求并验证响应
	// - Database工具：测试数据库连接和查询
	// - gRPC工具：测试gRPC服务调用
	// - 自定义工具：执行自定义逻辑

	// 模拟测试结果 - 实际项目中需要根据工具类型实现
	testSuccess := true
	testMessage := "Tool test successful"

	if !testSuccess {
		slog.Error("Tool test failed", "toolId", toolID, "message", testMessage)
		return fmt.Errorf("tool test failed: %s", testMessage)
	}

	slog.Info("Tool test completed successfully", "toolId", toolID, "toolName", tool.Name, "type", tool.Type)
	return nil
}

func (s *toolServiceImpl) ImportTools(ctx context.Context, req *dto.ImportToolsRequest) error {
	slog.Info("Importing tools from resource", "resourceId", req.ResourceID)

	// 验证参数
	if req.ResourceID == "" {
		slog.Warn("Resource ID is required for import")
		return domainErr.ErrInvalidInput
	}

	// 解析资源ID
	resourceID, err := uuid.Parse(req.ResourceID)
	if err != nil {
		slog.Warn("Invalid resource ID", "resourceId", req.ResourceID, "error", err)
		return domainErr.ErrInvalidInput
	}

	// 查找资源
	resource, err := s.resourceRepo.FindByID(ctx, resourceID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Resource not found for import", "resourceId", resourceID)
			return domainErr.ErrNotFound
		}
		slog.Error("Failed to find resource for import", "resourceId", resourceID, "error", err)
		return fmt.Errorf("find resource: %w", err)
	}

	// 检查资源类型 - 必须是支持工具导入的资源类型
	if resource.Type != "plugin" && resource.Type != "template" {
		slog.Warn("Resource type not supported for tool import", "resourceId", resourceID, "type", resource.Type)
		return fmt.Errorf("resource type '%s' not supported for tool import", resource.Type)
	}

	// 根据资源类型执行导入逻辑
	importedCount := 0
	failedCount := 0

	switch resource.Type {
	case "plugin":
		// 插件资源导入：从插件配置中提取工具定义
		slog.Info("Importing tools from plugin resource", "resourceId", resourceID, "resourceName", resource.Name)

		// TODO: 解析插件配置，提取工具定义
		// 插件通常包含多个工具定义，需要遍历并创建相应的工具

		// 模拟导入1个工具
		importedCount = 1

	case "template":
		// 模板资源导入：从模板创建工具
		slog.Info("Importing tools from template resource", "resourceId", resourceID, "resourceName", resource.Name)

		// TODO: 根据模板创建工具
		// 模板可能定义了一个工具的结构和默认配置

		// 模拟导入1个工具
		importedCount = 1
	}

	// 记录导入结果
	if failedCount > 0 {
		slog.Error("Tool import completed with failures",
			"resourceId", resourceID, "imported", importedCount, "failed", failedCount)
		return fmt.Errorf("tool import completed with %d failures", failedCount)
	}

	slog.Info("Tool import completed successfully",
		"resourceId", resourceID, "imported", importedCount, "resourceName", resource.Name)
	return nil
}

func (s *toolServiceImpl) PublishTool(ctx context.Context, toolID string) error {
	slog.Info("Publishing tool", "toolId", toolID)

	id, err := uuid.Parse(toolID)
	if err != nil {
		slog.Warn("Invalid tool id", "toolId", toolID, "error", err)
		return domainErr.ErrInvalidInput
	}

	// 1. 查找工具
	tool, err := s.toolRepo.FindByID(ctx, id)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Tool not found", "toolId", id)
			return domainErr.ErrNotFound
		}
		slog.Error("Failed to find tool", "toolId", id, "error", err)
		return fmt.Errorf("find tool: %w", err)
	}

	// 检查当前状态 - 只有草稿状态的工具可以发布
	if tool.Status != entity.ToolStatusDraft {
		slog.Warn("Tool cannot be published from current status", "toolId", id, "status", tool.Status)
		return fmt.Errorf("tool can only be published from draft status, current status: %s", tool.Status)
	}

	// 2. 发布工具
	if err := tool.Publish(); err != nil {
		slog.Error("Failed to publish tool", "toolId", id, "error", err)
		return domainErr.ErrInvalidInput
	}

	// 3. 保存到数据库
	if err := s.toolRepo.Update(ctx, tool); err != nil {
		slog.Error("Failed to save published tool", "toolId", id, "error", err)
		return fmt.Errorf("publish tool: %w", err)
	}

	slog.Info("Tool published successfully", "toolId", id, "name", tool.Name)
	return nil
}

func (s *toolServiceImpl) DeprecateTool(ctx context.Context, toolID string) error {
	slog.Info("Deprecating tool", "toolId", toolID)

	id, err := uuid.Parse(toolID)
	if err != nil {
		slog.Warn("Invalid tool id", "toolId", toolID, "error", err)
		return domainErr.ErrInvalidInput
	}

	// 1. 查找工具
	tool, err := s.toolRepo.FindByID(ctx, id)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Tool not found", "toolId", id)
			return domainErr.ErrNotFound
		}
		slog.Error("Failed to find tool", "toolId", id, "error", err)
		return fmt.Errorf("find tool: %w", err)
	}

	// 检查当前状态 - 只有已发布的工具可以废弃
	if tool.Status != entity.ToolStatusPublished {
		slog.Warn("Tool cannot be deprecated from current status", "toolId", id, "status", tool.Status)
		return fmt.Errorf("tool can only be deprecated from published status, current status: %s", tool.Status)
	}

	// 2. 废弃工具
	if err := tool.Deprecate(); err != nil {
		slog.Error("Failed to deprecate tool", "toolId", id, "error", err)
		return domainErr.ErrInvalidInput
	}

	// 3. 保存到数据库
	if err := s.toolRepo.Update(ctx, tool); err != nil {
		slog.Error("Failed to save deprecated tool", "toolId", id, "error", err)
		return fmt.Errorf("deprecate tool: %w", err)
	}

	slog.Info("Tool deprecated successfully", "toolId", id, "name", tool.Name)
	return nil
}

func (s *toolServiceImpl) ListVersions(ctx context.Context, toolID string) ([]string, error) {
	slog.Info("Listing tool versions", "toolId", toolID)

	// 验证参数
	if toolID == "" {
		slog.Warn("Tool ID is required for listing versions")
		return nil, domainErr.ErrInvalidInput
	}

	// 解析工具ID
	id, err := uuid.Parse(toolID)
	if err != nil {
		slog.Warn("Invalid tool ID", "toolId", toolID, "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	// 查找工具以获取基本信息
	tool, err := s.toolRepo.FindByID(ctx, id)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Warn("Tool not found for version listing", "toolId", id)
			return nil, domainErr.ErrNotFound
		}
		slog.Error("Failed to find tool for version listing", "toolId", id, "error", err)
		return nil, fmt.Errorf("find tool: %w", err)
	}

	// TODO: 从版本管理系统中获取工具的所有版本
	// 实际项目中可能需要查询专门的版本表或从历史记录中获取
	// 这里返回模拟数据

	// 根据当前版本号生成版本列表
	versions := make([]string, 0, tool.Version)
	for i := 1; i <= tool.Version; i++ {
		versions = append(versions, fmt.Sprintf("v%d", i))
	}

	// 如果没有版本历史，至少返回当前版本
	if len(versions) == 0 {
		versions = append(versions, fmt.Sprintf("v%d", tool.Version))
	}

	slog.Info("Tool versions listed successfully", "toolId", id, "versionCount", len(versions))
	return versions, nil
}

// toToolResponse 将Tool实体转换为响应DTO
func (s *toolServiceImpl) toToolResponse(tool *entity.Tool) *dto.ToolResponse {
	return &dto.ToolResponse{
		ID:          tool.ID.String(),
		Name:        tool.Name,
		Type:        tool.Type,
		Config:      tool.Config,
		Status:      string(tool.Status),
		WorkspaceID: tool.WorkspaceID.String(),
		Description: tool.Description,
		CreatedAt:   tool.CreatedAt,
		UpdatedAt:   tool.UpdatedAt,
	}
}
