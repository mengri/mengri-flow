package handler

import "github.com/gin-gonic/gin"

// IFlowHandler 流程处理器接口
type IFlowHandler interface {
	CreateFlow(c *gin.Context)
	ListFlows(c *gin.Context)
	GetFlow(c *gin.Context)
	UpdateFlow(c *gin.Context)
	DeleteFlow(c *gin.Context)
}

// Ensure FlowHandler implements IFlowHandler
var _ IFlowHandler = (*FlowHandler)(nil)
