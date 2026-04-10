package handler

import (
	"strconv"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/app/service"
	"mengri-flow/pkg/response"

	"github.com/gin-gonic/gin"
)

type ClusterHandlerImpl struct {
	service service.IClusterService `autowired:""`
}

// CreateCluster 创建集群
// @Summary 创建集群
// @Tags Cluster
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateClusterRequest true "创建集群请求"
// @Success 200 {object} response.Response{data=dto.ClusterResponse}
// @Failure 400 {object} response.Response
// @Router /clusters [post]
func (h *ClusterHandlerImpl) CreateCluster(c *gin.Context) {
	var req dto.CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	cluster, err := h.service.CreateCluster(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, cluster)
}

// ListClusters 获取集群列表
// @Summary 获取集群列表
// @Tags Cluster
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param environmentId query string false "环境ID"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=dto.ListClustersResponse}
// @Router /clusters [get]
func (h *ClusterHandlerImpl) ListClusters(c *gin.Context) {
	environmentID := c.Query("environmentId")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	clusters, err := h.service.ListClusters(c.Request.Context(), environmentID, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, clusters)
}

// GetClusterDetail 获取集群详情
// @Summary 获取集群详情
// @Tags Cluster
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "集群ID"
// @Success 200 {object} response.Response{data=dto.ClusterDetailResponse}
// @Failure 404 {object} response.Response
// @Router /clusters/{id} [get]
func (h *ClusterHandlerImpl) GetClusterDetail(c *gin.Context) {
	id := c.Param("id")

	cluster, err := h.service.GetClusterDetail(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, cluster)
}

// UpdateCluster 更新集群
// @Summary 更新集群
// @Tags Cluster
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "集群ID"
// @Param request body dto.UpdateClusterRequest true "更新集群请求"
// @Success 200 {object} response.Response{data=dto.ClusterResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /clusters/{id} [put]
func (h *ClusterHandlerImpl) UpdateCluster(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	cluster, err := h.service.UpdateCluster(c.Request.Context(), id, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, cluster)
}

// DeleteCluster 删除集群
// @Summary 删除集群
// @Tags Cluster
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "集群ID"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /clusters/{id} [delete]
func (h *ClusterHandlerImpl) DeleteCluster(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteCluster(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "cluster deleted successfully"})
}

// TestEtcdConnection 测试etcd连接
// @Summary 测试etcd连接
// @Tags Cluster
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "集群ID"
// @Param request body dto.TestEtcdConnectionRequest true "测试连接请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /clusters/{id}/test-etcd [post]
func (h *ClusterHandlerImpl) TestEtcdConnection(c *gin.Context) {
	var req dto.TestEtcdConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	result, err := h.service.TestEtcdConnection(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, result)
}
