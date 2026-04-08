package handler

import "github.com/gin-gonic/gin"

// IClusterHandler 集群处理器接口
type IClusterHandler interface {
	CreateCluster(c *gin.Context)
	ListClusters(c *gin.Context)
	GetClusterDetail(c *gin.Context)
	UpdateCluster(c *gin.Context)
	DeleteCluster(c *gin.Context)
	TestEtcdConnection(c *gin.Context)
}

// Ensure ClusterHandler implements IClusterHandler
var _ IClusterHandler = (*ClusterHandler)(nil)
