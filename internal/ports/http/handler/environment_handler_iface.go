package handler

import (
	"mengri-flow/pkg/autowire"

	"github.com/gin-gonic/gin"
)

// EnvironmentHandler 环境管理 HTTP 处理器接口
type IEnvironmentHandler interface {
	ListEnvironments(c *gin.Context)
	CreateEnvironment(c *gin.Context)
	GetEnvironment(c *gin.Context)
	UpdateEnvironment(c *gin.Context)
	DeleteEnvironment(c *gin.Context)
}

func init() {
	autowire.Auto(func() IEnvironmentHandler {
		return &EnvironmentHandlerImpl{}
	})
}
