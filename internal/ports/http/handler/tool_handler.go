package handler

import (
	"strconv"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/app/service"
	"mengri-flow/pkg/response"

	"github.com/gin-gonic/gin"
)

type ToolHandlerImpl struct {
	service service.IToolService `autowired:""`
}

// ListTools 获取工具列表
// @Summary 获取工具列表
// @Tags Tool
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspaceId query string true "工作空间ID"
// @Param type query string false "工具类型"
// @Param status query string false "状态"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=dto.ListToolsResponse}
// @Router /tools [get]
func (h *ToolHandlerImpl) ListTools(c *gin.Context) {
	workspaceID := c.Query("workspaceId")
	toolType := c.Query("type")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	req := &dto.ListToolsRequest{
		WorkspaceID: workspaceID,
		Type:        toolType,
		Status:      status,
		Page:        page,
		PageSize:    pageSize,
	}

	tools, err := h.service.ListTools(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, tools)
}

// CreateTool 创建工具
// @Summary 创建工具
// @Tags Tool
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateToolRequest true "创建工具请求"
// @Success 200 {object} response.Response{data=dto.ToolResponse}
// @Failure 400 {object} response.Response
// @Router /tools [post]
func (h *ToolHandlerImpl) CreateTool(c *gin.Context) {
	var req dto.CreateToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	tool, err := h.service.CreateTool(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, tool)
}

// GetTool 获取工具详情
// @Summary 获取工具详情
// @Tags Tool
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "工具ID"
// @Success 200 {object} response.Response{data=dto.ToolResponse}
// @Failure 404 {object} response.Response
// @Router /tools/{id} [get]
func (h *ToolHandlerImpl) GetTool(c *gin.Context) {
	id := c.Param("id")

	tool, err := h.service.GetTool(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, tool)
}

// UpdateTool 更新工具
// @Summary 更新工具
// @Tags Tool
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "工具ID"
// @Param request body dto.UpdateToolRequest true "更新工具请求"
// @Success 200 {object} response.Response{data=dto.ToolResponse}
// @Failure 400 {object} response.Response
// @Router /tools/{id} [put]
func (h *ToolHandlerImpl) UpdateTool(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	tool, err := h.service.UpdateTool(c.Request.Context(), id, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, tool)
}

// TestTool 测试工具
// @Summary 测试工具
// @Tags Tool
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.TestToolRequest true "测试工具请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /tools/test [post]
func (h *ToolHandlerImpl) TestTool(c *gin.Context) {
	var req dto.TestToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	if err := h.service.TestTool(c.Request.Context(), &req); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "tool tested successfully"})
}

// ImportTools 批量导入工具
// @Summary 批量导入工具
// @Tags Tool
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ImportToolsRequest true "导入工具请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /tools/import [post]
func (h *ToolHandlerImpl) ImportTools(c *gin.Context) {
	var req dto.ImportToolsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	if err := h.service.ImportTools(c.Request.Context(), &req); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "tools imported successfully"})
}

// PublishTool 发布工具
// @Summary 发布工具
// @Tags Tool
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "工具ID"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /tools/{id}/publish [post]
func (h *ToolHandlerImpl) PublishTool(c *gin.Context) {
	toolID := c.Param("id")

	if err := h.service.PublishTool(c.Request.Context(), toolID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "tool published successfully"})
}

// DeprecateTool 下线工具
// @Summary 下线工具
// @Tags Tool
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "工具ID"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /tools/{id}/deprecate [post]
func (h *ToolHandlerImpl) DeprecateTool(c *gin.Context) {
	toolID := c.Param("id")

	if err := h.service.DeprecateTool(c.Request.Context(), toolID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "tool deprecated successfully"})
}

// ListVersions 获取工具版本列表
// @Summary 获取工具版本列表
// @Tags Tool
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "工具ID"
// @Success 200 {object} response.Response{data=[]dto.ToolVersionResponse}
// @Failure 500 {object} response.Response
// @Router /tools/{id}/versions [get]
func (h *ToolHandlerImpl) ListVersions(c *gin.Context) {
	toolID := c.Param("id")

	versions, err := h.service.ListVersions(c.Request.Context(), toolID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"versions": versions})
}
