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

// ListRuns 获取运行记录列表
// @Summary 获取运行记录列表
// @Tags Run
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param flowId query string false "流程ID"
// @Param triggerId query string false "触发器ID"
// @Param status query string false "状态"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=dto.ListRunsResponse}
// @Router /runs [get]
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

// GetRunDetail 获取运行详情
// @Summary 获取运行详情
// @Tags Run
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "运行ID"
// @Success 200 {object} response.Response{data=dto.RunDetailResponse}
// @Failure 404 {object} response.Response
// @Router /runs/{id} [get]
func (h *RunHandlerImpl) GetRunDetail(c *gin.Context) {
	id := c.Param("id")

	run, err := h.service.GetRunDetail(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, run)
}

// GetExecutionTimeline 获取执行时间线
// @Summary 获取执行时间线
// @Tags Run
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "运行ID"
// @Success 200 {object} response.Response{data=dto.ExecutionTimelineResponse}
// @Failure 404 {object} response.Response
// @Router /runs/{id}/timeline [get]
func (h *RunHandlerImpl) GetExecutionTimeline(c *gin.Context) {
	id := c.Param("id")

	timeline, err := h.service.GetExecutionTimeline(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"timeline": timeline})
}

// RetryRun 重试运行
// @Summary 重试运行
// @Tags Run
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "运行ID"
// @Success 200 {object} response.Response{data=dto.RunResponse}
// @Failure 500 {object} response.Response
// @Router /runs/{id}/retry [post]
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

// GetRunStats 获取运行统计
// @Summary 获取运行统计
// @Tags Run
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=dto.RunStatsResponse}
// @Router /runs/stats [get]
func (h *RunHandlerImpl) GetRunStats(c *gin.Context) {
	stats, err := h.service.GetRunStats(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, stats)
}
