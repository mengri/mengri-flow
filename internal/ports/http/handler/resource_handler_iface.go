package handler

import (
	"mengri-flow/pkg/autowire"

	"github.com/gin-gonic/gin"
)

// ResourceHandler 资源管理 HTTP 处理器接口
type IResourceHandler interface {
	CreateResource(c *gin.Context)
	UpdateResource(c *gin.Context)
	DeleteResource(c *gin.Context)
	GetResource(c *gin.Context)
	ListResources(c *gin.Context)
	TestConnection(c *gin.Context)
}

func init() {
	autowire.Auto(func() IResourceHandler {
		return &ResourceHandlerImpl{}
	})
}
