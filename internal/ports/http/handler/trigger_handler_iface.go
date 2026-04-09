package handler

import (
	"mengri-flow/pkg/autowire"

	"github.com/gin-gonic/gin"
)

// TriggerHandler 触发器管理 HTTP 处理器接口
type ITriggerHandler interface {
	ListTriggers(c *gin.Context)
	CreateTrigger(c *gin.Context)
	GetTrigger(c *gin.Context)
	UpdateTrigger(c *gin.Context)
	DeleteTrigger(c *gin.Context)
	EnableTrigger(c *gin.Context)
	DisableTrigger(c *gin.Context)
	PublishToCluster(c *gin.Context)
}

func init() {
	autowire.Auto(func() ITriggerHandler {
		return &TriggerHandlerImpl{}
	})
}
