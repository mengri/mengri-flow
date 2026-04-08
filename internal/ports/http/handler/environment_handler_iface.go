package handler

import "github.com/gin-gonic/gin"

// IEnvironmentHandler 环境处理器接口
type IEnvironmentHandler interface {
	CreateEnvironment(c *gin.Context)
	ListEnvironments(c *gin.Context)
	GetEnvironment(c *gin.Context)
	UpdateEnvironment(c *gin.Context)
	DeleteEnvironment(c *gin.Context)
}

// Ensure EnvironmentHandler implements IEnvironmentHandler
var _ IEnvironmentHandler = (*EnvironmentHandler)(nil)
