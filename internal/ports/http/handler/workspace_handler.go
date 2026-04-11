package handler

import (
	"strconv"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/app/service"
	"mengri-flow/pkg/response"

	"github.com/gin-gonic/gin"
)

type WorkspaceHandlerImpl struct {
	service service.IWorkspaceService `autowired:""`
}

// CreateWorkspace 创建工作空间
// @Summary 创建工作空间
// @Tags Workspace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateWorkspaceRequest true "创建工作空间请求"
// @Success 200 {object} response.Response{data=dto.WorkspaceResponse}
// @Failure 400 {object} response.Response
// @Router /workspaces [post]
func (h *WorkspaceHandlerImpl) CreateWorkspace(c *gin.Context) {
	var req dto.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	accountID := c.GetString("accountID")
	workspace, err := h.service.CreateWorkspace(c.Request.Context(), &req, accountID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, workspace)
}

// ListWorkspaces 获取工作空间列表
// @Summary 获取工作空间列表
// @Tags Workspace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=dto.ListWorkspacesResponse}
// @Router /workspaces [get]
func (h *WorkspaceHandlerImpl) ListWorkspaces(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	accountID := c.GetString("accountID")
	workspaces, err := h.service.ListWorkspaces(c.Request.Context(), accountID, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, workspaces)
}

// GetWorkspace 获取工作空间详情
// @Summary 获取工作空间详情
// @Tags Workspace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "工作空间ID"
// @Success 200 {object} response.Response{data=dto.WorkspaceResponse}
// @Failure 404 {object} response.Response
// @Router /workspaces/{id} [get]
func (h *WorkspaceHandlerImpl) GetWorkspace(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	workspace, err := h.service.GetWorkspace(c.Request.Context(), id, accountID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, workspace)
}

// UpdateWorkspace 更新工作空间
// @Summary 更新工作空间
// @Tags Workspace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "工作空间ID"
// @Param request body dto.UpdateWorkspaceRequest true "更新工作空间请求"
// @Success 200 {object} response.Response{data=dto.WorkspaceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /workspaces/{id} [put]
func (h *WorkspaceHandlerImpl) UpdateWorkspace(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	var req dto.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	workspace, err := h.service.UpdateWorkspace(c.Request.Context(), id, &req, accountID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, workspace)
}

// DeleteWorkspace 删除工作空间
// @Summary 删除工作空间
// @Tags Workspace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "工作空间ID"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /workspaces/{id} [delete]
func (h *WorkspaceHandlerImpl) DeleteWorkspace(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	if err := h.service.DeleteWorkspace(c.Request.Context(), id, accountID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "workspace deleted successfully"})
}

// AddMember 添加工作空间成员
// @Summary 添加工作空间成员
// @Tags Workspace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "工作空间ID"
// @Param request body dto.AddWorkspaceMemberRequest true "添加成员请求"
// @Success 200 {object} response.Response{data=dto.WorkspaceMemberResponse}
// @Failure 400 {object} response.Response
// @Router /workspaces/{id}/members [post]
func (h *WorkspaceHandlerImpl) AddMember(c *gin.Context) {
	workspaceID := c.Param("id")
	accountID := c.GetString("accountID")

	var req dto.AddWorkspaceMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	member, err := h.service.AddMember(c.Request.Context(), workspaceID, &req, accountID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, member)
}

// RemoveMember 移除工作空间成员
// @Summary 移除工作空间成员
// @Tags Workspace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "工作空间ID"
// @Param userId path string true "成员ID"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /workspaces/{id}/members/{userId} [delete]
func (h *WorkspaceHandlerImpl) RemoveMember(c *gin.Context) {
	workspaceID := c.Param("id")
	memberID := c.Param("userId")
	accountID := c.GetString("accountID")

	if err := h.service.RemoveMember(c.Request.Context(), workspaceID, memberID, accountID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "member removed successfully"})
}

// ListMembers 获取工作空间成员列表
// @Summary 获取工作空间成员列表
// @Tags Workspace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "工作空间ID"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=dto.ListWorkspaceMembersResponse}
// @Router /workspaces/{id}/members [get]
func (h *WorkspaceHandlerImpl) ListMembers(c *gin.Context) {
	workspaceID := c.Param("id")
	accountID := c.GetString("accountID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	members, total, err := h.service.ListMembers(c.Request.Context(), workspaceID, accountID, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, &dto.ListWorkspaceMembersResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		List:     members,
	})
}
