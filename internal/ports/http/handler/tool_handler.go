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

func (h *ToolHandlerImpl) GetTool(c *gin.Context) {
	id := c.Param("id")

	tool, err := h.service.GetTool(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, tool)
}

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

func (h *ToolHandlerImpl) PublishTool(c *gin.Context) {
	toolID := c.Param("id")

	if err := h.service.PublishTool(c.Request.Context(), toolID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "tool published successfully"})
}

func (h *ToolHandlerImpl) DeprecateTool(c *gin.Context) {
	toolID := c.Param("id")

	if err := h.service.DeprecateTool(c.Request.Context(), toolID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "tool deprecated successfully"})
}

func (h *ToolHandlerImpl) ListVersions(c *gin.Context) {
	toolID := c.Param("id")

	versions, err := h.service.ListVersions(c.Request.Context(), toolID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"versions": versions})
}
