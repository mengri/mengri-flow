package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Heartbeat 心跳上报器
type Heartbeat struct {
	etcdClient  *clientv3.Client
	clusterID   string
	nodeID      string
	interval    time.Duration
	ttl         time.Duration
	stopChan    chan struct{}
	leaseID     clientv3.LeaseID
	mu          sync.Mutex
}

// NewHeartbeat 创建新的心跳上报器
func NewHeartbeat(etcdClient *clientv3.Client, clusterID, nodeID string) *Heartbeat {
	return &Heartbeat{
		etcdClient: etcdClient,
		clusterID:  clusterID,
		nodeID:     nodeID,
		interval:   5 * time.Second,
		ttl:        15 * time.Second,
		stopChan:   make(chan struct{}),
	}
}

// Start 启动心跳上报
func (h *Heartbeat) Start(ctx context.Context) {
	log.Printf("Starting heartbeat for node %s", h.nodeID)

	// 首次创建租约
	if err := h.createLease(ctx); err != nil {
		log.Printf("Failed to create initial lease: %v", err)
		return
	}

	// 启动心跳goroutine
	go func() {
		ticker := time.NewTicker(h.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := h.sendHeartbeat(ctx); err != nil {
					log.Printf("Heartbeat failed: %v", err)
					// 尝试重新创建租约
					if err := h.createLease(ctx); err != nil {
						log.Printf("Failed to recreate lease: %v", err)
					}
				}

			case <-h.stopChan:
				log.Printf("Heartbeat stopped for node %s", h.nodeID)
				return
			}
		}
	}()

	// 立即发送一次心跳
	if err := h.sendHeartbeat(ctx); err != nil {
		log.Printf("Initial heartbeat failed: %v", err)
	}
}

// Stop 停止心跳上报
func (h *Heartbeat) Stop() {
	close(h.stopChan)

	// 删除执行器状态
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("/clusters/%s/executors/%s", h.clusterID, h.nodeID)
	if _, err := h.etcdClient.Delete(ctx, key); err != nil {
		log.Printf("Failed to delete executor status: %v", err)
	}
}

// createLease 创建租约
func (h *Heartbeat) createLease(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	leaseResp, err := h.etcdClient.Grant(ctx, int64(h.ttl.Seconds()))
	if err != nil {
		return fmt.Errorf("failed to grant lease: %w", err)
	}

	h.leaseID = leaseResp.ID
	log.Printf("Created new lease %d for node %s", leaseResp.ID, h.nodeID)

	return nil
}

// sendHeartbeat 发送心跳
func (h *Heartbeat) sendHeartbeat(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 1. 续约租约
	_, err := h.etcdClient.KeepAliveOnce(ctx, h.leaseID)
	if err != nil {
		return fmt.Errorf("failed to keep alive lease: %w", err)
	}

	// 2. 构建状态信息
	status := h.buildStatus()

	// 3. 序列化
	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	// 4. 写入etcd
	key := fmt.Sprintf("/clusters/%s/executors/%s", h.clusterID, h.nodeID)
	_, err = h.etcdClient.Put(ctx, key, string(data), clientv3.WithLease(h.leaseID))
	if err != nil {
		return fmt.Errorf("failed to put status: %w", err)
	}

	log.Printf("Heartbeat sent for node %s", h.nodeID)

	return nil
}

// buildStatus 构建状态信息
func (h *Heartbeat) buildStatus() *ExecutorStatus {
	return &ExecutorStatus{
		NodeID:        h.nodeID,
		IP:            getLocalIP(),
		Hostname:      getHostname(),
		Version:       getVersion(),
		Status:        "active",
		CPUUsage:      getCPUUsage(),
		MemoryUsage:   getMemoryUsage(),
		GORoutines:    runtime.NumGoroutine(),
		LastHeartbeat: time.Now(),
	}
}

// ExecutorStatus 执行器状态
type ExecutorStatus struct {
	NodeID        string    `json:"nodeId"`
	IP            string    `json:"ip"`
	Hostname      string    `json:"hostname"`
	Version       string    `json:"version"`
	Status        string    `json:"status"`
	CPUUsage      float64   `json:"cpuUsage"`
	MemoryUsage   float64   `json:"memoryUsage"`
	GORoutines    int       `json:"goroutines"`
	RunningFlows  int       `json:"runningFlows"`
	LastHeartbeat time.Time `json:"lastHeartbeat"`
}

// getLocalIP 获取本地IP
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "unknown"
}

// getHostname 获取主机名
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// getVersion 获取版本号
func getVersion() string {
	return "1.0.0" // 应该从构建信息中获取
}

// getCPUUsage 获取CPU使用率（简化版）
func getCPUUsage() float64 {
	// 实际应该使用 gopsutil 等库获取真实CPU使用率
	return 0.0
}

// getMemoryUsage 获取内存使用率（简化版）
func getMemoryUsage() float64 {
	// 实际应该使用 gopsutil 等库获取真实内存使用率
	return 0.0
}

// getRunningFlows 获取正在执行的流程数（简化版）
func getRunningFlows() int {
	// 实际应该返回当前正在执行的流程数量
	return 0
}
