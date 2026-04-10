package handler

import (
	"strconv"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/app/service"
	"mengri-flow/pkg/response"

	"github.com/gin-gonic/gin"
)

type FlowHandlerImpl struct {
	service service.IFlowService `autowired:""`
}

// CreateFlow 创建流程
// @Summary 创建流程
// @Tags Flow
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateFlowRequest true "创建流程请求"
// @Success 200 {object} response.Response{data=dto.FlowResponse}
// @Failure 400 {object} response.Response
// @Router /flows [post]
func (h *FlowHandlerImpl) CreateFlow(c *gin.Context) {
	var req dto.CreateFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	accountID := c.GetString("accountID")
	flow, err := h.service.CreateFlow(c.Request.Context(), &req, accountID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, flow)
}

// ListFlows 获取流程列表
// @Summary 获取流程列表
// @Tags Flow
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspaceId query string true "工作空间ID"
// @Param status query string false "状态"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=dto.ListFlowsResponse}
// @Router /flows [get]
func (h *FlowHandlerImpl) ListFlows(c *gin.Context) {
	workspaceID := c.Query("workspaceId")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	req := &dto.ListFlowsRequest{
		WorkspaceID: workspaceID,
		Status:      status,
		Page:        page,
		PageSize:    pageSize,
	}

	flows, err := h.service.ListFlows(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, flows)
}

// GetFlow 获取流程详情
// @Summary 获取流程详情
// @Tags Flow
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "流程ID"
// @Success 200 {object} response.Response{data=dto.FlowResponse}
// @Failure 404 {object} response.Response
// @Router /flows/{id} [get]
func (h *FlowHandlerImpl) GetFlow(c *gin.Context) {
	id := c.Param("id")

	flow, err := h.service.GetFlow(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, flow)
}

// UpdateFlow 更新流程
// @Summary 更新流程
// @Tags Flow
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "流程ID"
// @Param request body dto.UpdateFlowRequest true "更新流程请求"
// @Success 200 {object} response.Response{data=dto.FlowResponse}
// @Failure 400 {object} response.Response
// @Router /flows/{id} [put]
func (h *FlowHandlerImpl) UpdateFlow(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	var req dto.UpdateFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	flow, err := h.service.UpdateFlow(c.Request.Context(), id, &req, accountID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, flow)
}

// DeleteFlow 删除流程
// @Summary 删除流程
// @Tags Flow
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "流程ID"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /flows/{id} [delete]
func (h *FlowHandlerImpl) DeleteFlow(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	if err := h.service.DeleteFlow(c.Request.Context(), id, accountID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "flow deleted successfully"})
}

// TestFlow 测试流程
// @Summary 测试流程
// @Tags Flow
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.TestFlowRequest true "测试流程请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /flows/test [post]
func (h *FlowHandlerImpl) TestFlow(c *gin.Context) {
	var req dto.TestFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	if err := h.service.TestFlow(c.Request.Context(), &req); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "flow test completed"})
}

// PublishFlow 发布流程
// @Summary 发布流程
// @Tags Flow
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "流程ID"
// @Param request body dto.PublishFlowRequest true "发布流程请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /flows/{id}/publish [post]
func (h *FlowHandlerImpl) PublishFlow(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	var req dto.PublishFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	if err := h.service.PublishFlow(c.Request.Context(), id, req.ClusterID, accountID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "flow published successfully"})
}

// ListVersions 获取流程版本列表
// @Summary 获取流程版本列表
// @Tags Flow
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "流程ID"
// @Success 200 {object} response.Response{data=[]dto.FlowVersionResponse}
// @Router /flows/{id}/versions [get]
func (h *FlowHandlerImpl) ListVersions(c *gin.Context) {
	id := c.Param("id")

	versions, err := h.service.ListVersions(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"versions": versions})
}

// RollbackVersion 回滚流程版本
// @Summary 回滚流程版本
// @Tags Flow
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "流程ID"
// @Param request body dto.RollbackFlowRequest true "回滚请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /flows/{id}/rollback [post]
func (h *FlowHandlerImpl) RollbackVersion(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	var req dto.RollbackFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	if err := h.service.RollbackVersion(c.Request.Context(), id, req.Version, accountID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "flow rolled back successfully"})
}
