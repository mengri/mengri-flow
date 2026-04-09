package handler

import (
	"mengri-flow/pkg/autowire"

	"github.com/gin-gonic/gin"
)

// ClusterHandler 集群管理 HTTP 处理器接口
type IClusterHandler interface {
	ListClusters(c *gin.Context)
	CreateCluster(c *gin.Context)
	GetClusterDetail(c *gin.Context)
	UpdateCluster(c *gin.Context)
	DeleteCluster(c *gin.Context)
	TestEtcdConnection(c *gin.Context)
}

func init() {
	autowire.Auto(func() IClusterHandler {
		return &ClusterHandlerImpl{}
	})
}
