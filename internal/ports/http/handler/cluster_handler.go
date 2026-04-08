package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/app/service"
	"mengri-flow/pkg/response"
)

type ClusterHandler struct {
	service service.IClusterService
}

func NewClusterHandler(svc service.IClusterService) *ClusterHandler {
	return &ClusterHandler{
		service: svc,
	}
}

func (h *ClusterHandler) CreateCluster(c *gin.Context) {
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

func (h *ClusterHandler) ListClusters(c *gin.Context) {
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

func (h *ClusterHandler) GetClusterDetail(c *gin.Context) {
	id := c.Param("id")

	cluster, err := h.service.GetClusterDetail(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, cluster)
}

func (h *ClusterHandler) UpdateCluster(c *gin.Context) {
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

func (h *ClusterHandler) DeleteCluster(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteCluster(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "cluster deleted successfully"})
}

func (h *ClusterHandler) TestEtcdConnection(c *gin.Context) {
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
