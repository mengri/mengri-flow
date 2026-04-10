package handler

import (
	"strconv"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/app/service"
	"mengri-flow/pkg/response"

	"github.com/gin-gonic/gin"
)

type EnvironmentHandlerImpl struct {
	service service.IEnvironmentService `autowired:""`
}

// CreateEnvironment 创建环境
// @Summary 创建环境
// @Tags Environment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateEnvironmentRequest true "创建环境请求"
// @Success 200 {object} response.Response{data=dto.EnvironmentResponse}
// @Failure 400 {object} response.Response
// @Router /environments [post]
func (h *EnvironmentHandlerImpl) CreateEnvironment(c *gin.Context) {
	var req dto.CreateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	env, err := h.service.CreateEnvironment(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, env)
}

// ListEnvironments 获取环境列表
// @Summary 获取环境列表
// @Tags Environment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=dto.ListEnvironmentsResponse}
// @Router /environments [get]
func (h *EnvironmentHandlerImpl) ListEnvironments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	envs, err := h.service.ListEnvironments(c.Request.Context(), page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, envs)
}

// GetEnvironment 获取环境详情
// @Summary 获取环境详情
// @Tags Environment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "环境ID"
// @Success 200 {object} response.Response{data=dto.EnvironmentResponse}
// @Failure 404 {object} response.Response
// @Router /environments/{id} [get]
func (h *EnvironmentHandlerImpl) GetEnvironment(c *gin.Context) {
	id := c.Param("id")

	env, err := h.service.GetEnvironment(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, env)
}

// UpdateEnvironment 更新环境
// @Summary 更新环境
// @Tags Environment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "环境ID"
// @Param request body dto.UpdateEnvironmentRequest true "更新环境请求"
// @Success 200 {object} response.Response{data=dto.EnvironmentResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /environments/{id} [put]
func (h *EnvironmentHandlerImpl) UpdateEnvironment(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	env, err := h.service.UpdateEnvironment(c.Request.Context(), id, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, env)
}

// DeleteEnvironment 删除环境
// @Summary 删除环境
// @Tags Environment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "环境ID"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /environments/{id} [delete]
func (h *EnvironmentHandlerImpl) DeleteEnvironment(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteEnvironment(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "environment deleted successfully"})
}
