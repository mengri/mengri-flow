package handler

import (
	"strconv"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/app/service"
	"mengri-flow/pkg/response"

	"github.com/gin-gonic/gin"
)

type TriggerHandlerImpl struct {
	service service.ITriggerService `autowired:""`
}

// ListTriggers 获取触发器列表
// @Summary 获取触发器列表
// @Tags Trigger
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param flowId query string false "流程ID"
// @Param status query string false "状态"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=dto.ListTriggersResponse}
// @Router /triggers [get]
func (h *TriggerHandlerImpl) ListTriggers(c *gin.Context) {
	flowID := c.Query("flowId")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	req := &dto.ListTriggersRequest{
		FlowID:   flowID,
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	}

	triggers, err := h.service.ListTriggers(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, triggers)
}

// CreateTrigger 创建触发器
// @Summary 创建触发器
// @Tags Trigger
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateTriggerRequest true "创建触发器请求"
// @Success 200 {object} response.Response{data=dto.TriggerResponse}
// @Failure 400 {object} response.Response
// @Router /triggers [post]
func (h *TriggerHandlerImpl) CreateTrigger(c *gin.Context) {
	var req dto.CreateTriggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	accountID := c.GetString("accountID")
	trigger, err := h.service.CreateTrigger(c.Request.Context(), &req, accountID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, trigger)
}

// GetTrigger 获取触发器详情
// @Summary 获取触发器详情
// @Tags Trigger
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "触发器ID"
// @Success 200 {object} response.Response{data=dto.TriggerResponse}
// @Failure 404 {object} response.Response
// @Router /triggers/{id} [get]
func (h *TriggerHandlerImpl) GetTrigger(c *gin.Context) {
	id := c.Param("id")

	trigger, err := h.service.GetTrigger(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, trigger)
}

// UpdateTrigger 更新触发器
// @Summary 更新触发器
// @Tags Trigger
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "触发器ID"
// @Param request body dto.UpdateTriggerRequest true "更新触发器请求"
// @Success 200 {object} response.Response{data=dto.TriggerResponse}
// @Failure 400 {object} response.Response
// @Router /triggers/{id} [put]
func (h *TriggerHandlerImpl) UpdateTrigger(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	var req dto.UpdateTriggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	trigger, err := h.service.UpdateTrigger(c.Request.Context(), id, &req, accountID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, trigger)
}

// DeleteTrigger 删除触发器
// @Summary 删除触发器
// @Tags Trigger
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "触发器ID"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /triggers/{id} [delete]
func (h *TriggerHandlerImpl) DeleteTrigger(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	if err := h.service.DeleteTrigger(c.Request.Context(), id, accountID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "trigger deleted successfully"})
}

// EnableTrigger 启用触发器
// @Summary 启用触发器
// @Tags Trigger
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "触发器ID"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /triggers/{id}/enable [post]
func (h *TriggerHandlerImpl) EnableTrigger(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	if err := h.service.EnableTrigger(c.Request.Context(), id, accountID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "trigger enabled successfully"})
}

// DisableTrigger 禁用触发器
// @Summary 禁用触发器
// @Tags Trigger
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "触发器ID"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /triggers/{id}/disable [post]
func (h *TriggerHandlerImpl) DisableTrigger(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	if err := h.service.DisableTrigger(c.Request.Context(), id, accountID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "trigger disabled successfully"})
}

// PublishToCluster 发布触发器到集群
// @Summary 发布触发器到集群
// @Tags Trigger
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "触发器ID"
// @Param request body dto.PublishTriggerRequest true "发布请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /triggers/{id}/publish [post]
func (h *TriggerHandlerImpl) PublishToCluster(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	var req dto.PublishTriggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	if err := h.service.PublishToCluster(c.Request.Context(), id, req.ClusterID, accountID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "trigger published successfully"})
}
