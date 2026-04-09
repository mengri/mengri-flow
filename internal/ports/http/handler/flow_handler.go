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

func (h *FlowHandlerImpl) ListFlows(c *gin.Context) {
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

func (h *FlowHandlerImpl) CreateFlow(c *gin.Context) {
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

func (h *FlowHandlerImpl) GetFlow(c *gin.Context) {
	id := c.Param("id")

	flow, err := h.service.GetFlow(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, flow)
}

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

func (h *FlowHandlerImpl) DeleteFlow(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	if err := h.service.DeleteFlow(c.Request.Context(), id, accountID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "flow deleted successfully"})
}

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

func (h *FlowHandlerImpl) ListVersions(c *gin.Context) {
	id := c.Param("id")

	versions, err := h.service.ListVersions(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"versions": versions})
}

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
