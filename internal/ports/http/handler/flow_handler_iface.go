package handler

import (
	"mengri-flow/pkg/autowire"

	"github.com/gin-gonic/gin"
)

// FlowHandler 流程管理 HTTP 处理器接口
type IFlowHandler interface {
	ListFlows(c *gin.Context)
	CreateFlow(c *gin.Context)
	GetFlow(c *gin.Context)
	UpdateFlow(c *gin.Context)
	DeleteFlow(c *gin.Context)
	TestFlow(c *gin.Context)
	PublishFlow(c *gin.Context)
	ListVersions(c *gin.Context)
	RollbackVersion(c *gin.Context)
}

func init() {
	autowire.Auto(func() IFlowHandler {
		return &FlowHandlerImpl{}
	})
}
