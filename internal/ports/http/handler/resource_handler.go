package handler

import (
	"net/http"
	"strconv"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/app/service"
	"mengri-flow/pkg/response"

	"github.com/gin-gonic/gin"
)

type ResourceHandlerImpl struct {
	resourceService service.IResourceService `autowired:""`
}

// CreateResource 创建资源
// @Summary 创建资源
// @Tags Resource
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateResourceRequest true "创建资源请求"
// @Success 200 {object} response.Response{data=dto.ResourceResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /resources [post]
func (h *ResourceHandlerImpl) CreateResource(c *gin.Context) {
	var req dto.CreateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	resource, err := h.resourceService.CreateResource(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to create resource", err)
		return
	}

	resp := &dto.ResourceResponse{
		ID:          resource.ID.String(),
		Name:        resource.Name,
		Type:        string(resource.Type),
		Config:      resource.Config,
		Status:      string(resource.Status),
		WorkspaceID: resource.WorkspaceID.String(),
		Description: resource.Description,
		CreatedAt:   resource.CreatedAt,
		UpdatedAt:   resource.UpdatedAt,
	}

	response.Success(c, resp)
}

// UpdateResource 更新资源
// @Summary 更新资源
// @Tags Resource
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "资源ID"
// @Param request body dto.UpdateResourceRequest true "更新资源请求"
// @Success 200 {object} response.Response{data=dto.ResourceResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /resources/{id} [put]
func (h *ResourceHandlerImpl) UpdateResource(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	resource, err := h.resourceService.UpdateResource(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to update resource", err)
		return
	}

	resp := &dto.ResourceResponse{
		ID:          resource.ID.String(),
		Name:        resource.Name,
		Type:        string(resource.Type),
		Config:      resource.Config,
		Status:      string(resource.Status),
		WorkspaceID: resource.WorkspaceID.String(),
		Description: resource.Description,
		CreatedAt:   resource.CreatedAt,
		UpdatedAt:   resource.UpdatedAt,
	}

	response.Success(c, resp)
}

// DeleteResource 删除资源
// @Summary 删除资源
// @Tags Resource
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "资源ID"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /resources/{id} [delete]
func (h *ResourceHandlerImpl) DeleteResource(c *gin.Context) {
	id := c.Param("id")

	if err := h.resourceService.DeleteResource(c.Request.Context(), id); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to delete resource", err)
		return
	}

	response.Success(c, gin.H{"message": "resource deleted"})
}

// GetResource 获取资源详情
// @Summary 获取资源详情
// @Tags Resource
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "资源ID"
// @Success 200 {object} response.Response{data=dto.ResourceResponse}
// @Failure 404 {object} response.Response
// @Router /resources/{id} [get]
func (h *ResourceHandlerImpl) GetResource(c *gin.Context) {
	id := c.Param("id")

	resource, err := h.resourceService.GetResource(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "resource not found", err)
		return
	}

	resp := &dto.ResourceResponse{
		ID:          resource.ID.String(),
		Name:        resource.Name,
		Type:        string(resource.Type),
		Config:      resource.Config,
		Status:      string(resource.Status),
		WorkspaceID: resource.WorkspaceID.String(),
		Description: resource.Description,
		CreatedAt:   resource.CreatedAt,
		UpdatedAt:   resource.UpdatedAt,
	}

	response.Success(c, resp)
}

// ListResources 获取资源列表
// @Summary 获取资源列表
// @Tags Resource
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspaceId query string true "工作空间ID"
// @Param type query string false "资源类型"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=dto.ListResourcesResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /resources [get]
func (h *ResourceHandlerImpl) ListResources(c *gin.Context) {
	workspaceID := c.Query("workspaceId")
	if workspaceID == "" {
		response.Error(c, http.StatusBadRequest, "workspaceId is required", nil)
		return
	}

	resourceType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	resources, total, err := h.resourceService.ListResources(c.Request.Context(), workspaceID, resourceType, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list resources", err)
		return
	}

	list := make([]dto.ResourceResponse, len(resources))
	for i, resource := range resources {
		list[i] = dto.ResourceResponse{
			ID:          resource.ID.String(),
			Name:        resource.Name,
			Type:        string(resource.Type),
			Config:      resource.Config,
			Status:      string(resource.Status),
			WorkspaceID: resource.WorkspaceID.String(),
			Description: resource.Description,
			CreatedAt:   resource.CreatedAt,
			UpdatedAt:   resource.UpdatedAt,
		}
	}

	resp := &dto.ListResourcesResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		List:     list,
	}

	response.Success(c, resp)
}

// TestConnection 测试资源连接
// @Summary 测试资源连接
// @Tags Resource
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.TestConnectionRequest true "测试连接请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /resources/test-connection [post]
func (h *ResourceHandlerImpl) TestConnection(c *gin.Context) {
	var req dto.TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	if err := h.resourceService.TestConnection(c.Request.Context(), &req); err != nil {
		response.Error(c, http.StatusBadRequest, "connection test failed", err)
		return
	}

	response.Success(c, gin.H{"message": "connection successful"})
}
