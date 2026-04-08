package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/app/service"
	"mengri-flow/pkg/response"
)

type RunHandler struct {
	service service.IRunService
}

func NewRunHandler(svc service.IRunService) *RunHandler {
	return &RunHandler{
		service: svc,
	}
}

func (h *RunHandler) ListRuns(c *gin.Context) {
	flowID := c.Query("flowId")
	triggerID := c.Query("triggerId")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	req := &dto.ListRunsRequest{
		FlowID:    flowID,
		TriggerID: triggerID,
		Status:    status,
		Page:      page,
		PageSize:  pageSize,
	}

	runs, err := h.service.ListRuns(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, runs)
}

func (h *RunHandler) GetRunDetail(c *gin.Context) {
	id := c.Param("id")

	run, err := h.service.GetRunDetail(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, run)
}

func (h *RunHandler) GetExecutionTimeline(c *gin.Context) {
	id := c.Param("id")

	timeline, err := h.service.GetExecutionTimeline(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"timeline": timeline})
}

func (h *RunHandler) RetryRun(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	req := &dto.RetryRunRequest{
		RunID: id,
	}

	run, err := h.service.RetryRun(c.Request.Context(), id, accountID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, run)
}

func (h *RunHandler) GetRunStats(c *gin.Context) {
	stats, err := h.service.GetRunStats(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, stats)
}
