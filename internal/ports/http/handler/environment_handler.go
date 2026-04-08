package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/app/service"
	"mengri-flow/pkg/response"
)

type EnvironmentHandler struct {
	service service.IEnvironmentService
}

func NewEnvironmentHandler(svc service.IEnvironmentService) *EnvironmentHandler {
	return &EnvironmentHandler{
		service: svc,
	}
}

func (h *EnvironmentHandler) CreateEnvironment(c *gin.Context) {
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

func (h *EnvironmentHandler) ListEnvironments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	envs, err := h.service.ListEnvironments(c.Request.Context(), page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, envs)
}

func (h *EnvironmentHandler) GetEnvironment(c *gin.Context) {
	id := c.Param("id")

	env, err := h.service.GetEnvironment(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, env)
}

func (h *EnvironmentHandler) UpdateEnvironment(c *gin.Context) {
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

func (h *EnvironmentHandler) DeleteEnvironment(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteEnvironment(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "environment deleted successfully"})
}
