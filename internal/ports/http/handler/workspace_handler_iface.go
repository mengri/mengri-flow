package handler

import "github.com/gin-gonic/gin"

// IWorkspaceHandler 工作空间处理器接口
type IWorkspaceHandler interface {
	CreateWorkspace(c *gin.Context)
	ListWorkspaces(c *gin.Context)
	GetWorkspace(c *gin.Context)
	UpdateWorkspace(c *gin.Context)
	DeleteWorkspace(c *gin.Context)
}

// Ensure WorkspaceHandler implements IWorkspaceHandler
var _ IWorkspaceHandler = (*WorkspaceHandler)(nil)
