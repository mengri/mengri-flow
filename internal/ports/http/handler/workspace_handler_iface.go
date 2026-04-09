package handler

import (
	"mengri-flow/pkg/autowire"

	"github.com/gin-gonic/gin"
)

// WorkspaceHandler 工作空间管理 HTTP 处理器接口
type IWorkspaceHandler interface {
	ListWorkspaces(c *gin.Context)
	CreateWorkspace(c *gin.Context)
	GetWorkspace(c *gin.Context)
	UpdateWorkspace(c *gin.Context)
	DeleteWorkspace(c *gin.Context)
	AddMember(c *gin.Context)
	RemoveMember(c *gin.Context)
}

func init() {
	autowire.Auto(func() IWorkspaceHandler {
		return &WorkspaceHandlerImpl{}
	})
}
