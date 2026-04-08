package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/app/service"
	"mengri-flow/pkg/response"
)

type TriggerHandler struct {
	service service.ITriggerService
}

func NewTriggerHandler(svc service.ITriggerService) *TriggerHandler {
	return &TriggerHandler{
		service: svc,
	}
}

func (h *TriggerHandler) CreateTrigger(c *gin.Context) {
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

func (h *TriggerHandler) ListTriggers(c *gin.Context) {
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

func (h *TriggerHandler) GetTrigger(c *gin.Context) {
	id := c.Param("id")

	trigger, err := h.service.GetTrigger(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, trigger)
}

func (h *TriggerHandler) UpdateTrigger(c *gin.Context) {
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

func (h *TriggerHandler) DeleteTrigger(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	if err := h.service.DeleteTrigger(c.Request.Context(), id, accountID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "trigger deleted successfully"})
}

func (h *TriggerHandler) EnableTrigger(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	if err := h.service.EnableTrigger(c.Request.Context(), id, accountID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "trigger enabled successfully"})
}

func (h *TriggerHandler) DisableTrigger(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	if err := h.service.DisableTrigger(c.Request.Context(), id, accountID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "trigger disabled successfully"})
}

func (h *TriggerHandler) PublishToCluster(c *gin.Context) {
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
