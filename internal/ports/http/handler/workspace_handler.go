package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/app/service"
	"mengri-flow/pkg/response"
)

type WorkspaceHandler struct {
	service service.IWorkspaceService
}

func NewWorkspaceHandler(svc service.IWorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{
		service: svc,
	}
}

func (h *WorkspaceHandler) CreateWorkspace(c *gin.Context) {
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

func (h *WorkspaceHandler) ListWorkspaces(c *gin.Context) {
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

func (h *WorkspaceHandler) GetWorkspace(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	workspace, err := h.service.GetWorkspace(c.Request.Context(), id, accountID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, workspace)
}

func (h *WorkspaceHandler) UpdateWorkspace(c *gin.Context) {
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

func (h *WorkspaceHandler) DeleteWorkspace(c *gin.Context) {
	id := c.Param("id")
	accountID := c.GetString("accountID")

	if err := h.service.DeleteWorkspace(c.Request.Context(), id, accountID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "workspace deleted successfully"})
}

func (h *WorkspaceHandler) AddMember(c *gin.Context) {
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

func (h *WorkspaceHandler) RemoveMember(c *gin.Context) {
	workspaceID := c.Param("id")
	memberID := c.Param("userId")
	accountID := c.GetString("accountID")

	if err := h.service.RemoveMember(c.Request.Context(), workspaceID, memberID, accountID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "member removed successfully"})
}
