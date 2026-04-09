package handler

import (
	"strconv"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/app/service"
	"mengri-flow/pkg/response"

	"github.com/gin-gonic/gin"
)

type RunHandlerImpl struct {
	service service.IRunService `autowired:""`
}

func (h *RunHandlerImpl) ListRuns(c *gin.Context) {
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

func (h *RunHandlerImpl) GetRunDetail(c *gin.Context) {
	id := c.Param("id")

	run, err := h.service.GetRunDetail(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, run)
}

func (h *RunHandlerImpl) GetExecutionTimeline(c *gin.Context) {
	id := c.Param("id")

	timeline, err := h.service.GetExecutionTimeline(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"timeline": timeline})
}

func (h *RunHandlerImpl) RetryRun(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	run, err := h.service.RetryRun(c.Request.Context(), id, accountID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, run)
}

func (h *RunHandlerImpl) GetRunStats(c *gin.Context) {
	stats, err := h.service.GetRunStats(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, stats)
}
