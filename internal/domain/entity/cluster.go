package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ClusterStatus 集群状态
type ClusterStatus string

const (
	ClusterStatusActive   ClusterStatus = "active"
	ClusterStatusInactive ClusterStatus = "inactive"
	ClusterStatusError    ClusterStatus = "error"
)

// Cluster 表示一个集群
type Cluster struct {
	ID            uuid.UUID
	Name          string
	EnvironmentID uuid.UUID
	EtcdEndpoints []string
	EtcdUsername  string
	EtcdPassword  string // 加密存储
	EtcdPrefix    string
	Description   string
	Status        ClusterStatus
	ExecutorCount int
	LastHeartbeat *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (c *Cluster) IsActive() bool {
	return c.Status == ClusterStatusActive
}

// NewCluster 创建一个新的集群
func NewCluster(name string, environmentID uuid.UUID, endpoints []string, username, password, prefix, description string) (*Cluster, error) {
	if name == "" {
		return nil, fmt.Errorf("cluster name cannot be empty")
	}

	if len(endpoints) == 0 {
		return nil, fmt.Errorf("etcd endpoints cannot be empty")
	}

	now := time.Now()
	return &Cluster{
		ID:            uuid.New(),
		Name:          name,
		EnvironmentID: environmentID,
		EtcdEndpoints: endpoints,
		EtcdUsername:  username,
		EtcdPassword:  password, // 应在服务层加密
		EtcdPrefix:    prefix,
		Description:   description,
		Status:        ClusterStatusActive,
		ExecutorCount: 0,
		LastHeartbeat: nil,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// Update 更新集群信息
func (c *Cluster) Update(name, description string) error {
	if name == "" {
		return fmt.Errorf("cluster name cannot be empty")
	}

	c.Name = name
	c.Description = description
	c.UpdatedAt = time.Now()
	return nil
}

// UpdateExecutorStatus 更新执行器状态
func (c *Cluster) UpdateExecutorStatus(count int, lastHeartbeat time.Time) {
	c.ExecutorCount = count
	c.LastHeartbeat = &lastHeartbeat
	c.UpdatedAt = time.Now()
}

// UpdateStatus 更新集群状态
func (c *Cluster) UpdateStatus(status ClusterStatus) {
	c.Status = status
	c.UpdatedAt = time.Now()
}
