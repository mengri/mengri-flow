package dto

import (
	"time"
)

type CreateClusterRequest struct {
	Name          string `json:"name" binding:"required,max=100"`
	Description   string `json:"description" binding:"max=500"`
	EnvironmentID string `json:"environmentId" binding:"required"`
	EtcdEndpoints string `json:"etcdEndpoints" binding:"required"`
	EtcdUsername  string `json:"etcdUsername" binding:"omitempty"`
	EtcdPassword  string `json:"etcdPassword" binding:"omitempty"`
}

type UpdateClusterRequest struct {
	Name        string `json:"name" binding:"omitempty,max=100"`
	Description string `json:"description" binding:"omitempty,max=500"`
}

type ClusterResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	EnvironmentID string    `json:"environmentId"`
	EtcdEndpoints string    `json:"etcdEndpoints"`
	EtcdUsername  string    `json:"etcdUsername"`
	Status        string    `json:"status"`
	NodeCount     int       `json:"nodeCount"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type ClusterDetailResponse struct {
	ClusterResponse
	ExecutorCount int `json:"executorCount"`
	ActiveFlows   int `json:"activeFlows"`
}

type TestEtcdConnectionRequest struct {
	Endpoints string `json:"endpoints" binding:"required"`
	Username  string `json:"username" binding:"omitempty"`
	Password  string `json:"password" binding:"omitempty"`
}

type TestEtcdConnectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ListClustersRequest struct {
	EnvironmentID string `json:"environmentId" binding:"omitempty"`
	Page          int    `json:"page" binding:"min=1"`
	PageSize      int    `json:"pageSize" binding:"min=1,max=100"`
}

type ListClustersResponse struct {
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
	List     []ClusterResponse `json:"list"`
}
