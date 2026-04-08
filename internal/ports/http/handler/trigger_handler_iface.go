package handler

import "github.com/gin-gonic/gin"

// ITriggerHandler 触发器处理器接口
type ITriggerHandler interface {
	CreateTrigger(c *gin.Context)
	ListTriggers(c *gin.Context)
	GetTrigger(c *gin.Context)
	UpdateTrigger(c *gin.Context)
	DeleteTrigger(c *gin.Context)
}

// Ensure TriggerHandler implements ITriggerHandler
var _ ITriggerHandler = (*TriggerHandler)(nil)
