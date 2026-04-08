package handler

import "github.com/gin-gonic/gin"

// IToolHandler 工具处理器接口
type IToolHandler interface {
	ListTools(c *gin.Context)
	CreateTool(c *gin.Context)
	GetTool(c *gin.Context)
	UpdateTool(c *gin.Context)
	TestTool(c *gin.Context)
	ImportTools(c *gin.Context)
	PublishTool(c *gin.Context)
	DeprecateTool(c *gin.Context)
	ListVersions(c *gin.Context)
}

// Ensure ToolHandler implements IToolHandler
var _ IToolHandler = (*ToolHandler)(nil)
