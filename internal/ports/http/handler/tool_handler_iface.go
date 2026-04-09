package handler

import (
	"mengri-flow/pkg/autowire"

	"github.com/gin-gonic/gin"
)

// ToolHandler 工具管理 HTTP 处理器接口
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

func init() {
	autowire.Auto(func() IToolHandler {
		return &ToolHandlerImpl{}
	})
}
