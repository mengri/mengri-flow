package executor

import (
	"context"
	"fmt"
	"log"
	"time"

	"mengri-flow/internal/infra/config"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Executor 执行器主结构体
type Executor struct {
	config         *config.ExecutorConfig
	etcdClient     *clientv3.Client
	triggerManager *TriggerManager
	flowEngine     *FlowEngine
	heartbeat      *Heartbeat
	configWatcher  *ConfigWatcher
}

// NewExecutor 创建新的执行器
func NewExecutor(cfg *config.ExecutorConfig) *Executor {
	return &Executor{
		config: cfg,
	}
}

// Start 启动执行器
func (e *Executor) Start(ctx context.Context) error {
	log.Printf("Starting executor %s for cluster %s", e.config.NodeID, e.config.ClusterID)

	// 1. 连接etcd
	if err := e.connectEtcd(); err != nil {
		return fmt.Errorf("failed to connect etcd: %w", err)
	}

	// 2. 初始化组件
	e.flowEngine = NewFlowEngine(e.config.NodeID)
	e.triggerManager = NewTriggerManager(e.flowEngine, e.config.NodeID)
	e.heartbeat = NewHeartbeat(e.etcdClient, e.config.ClusterID, e.config.NodeID)
	e.configWatcher = NewConfigWatcher(e.etcdClient, e.config.ClusterID, e.triggerManager)

	// 3. 启动心跳
	e.heartbeat.Start(ctx)

	// 4. 加载当前配置
	if err := e.loadInitialConfig(ctx); err != nil {
		log.Printf("Warning: failed to load initial config: %v", err)
	}

	// 5. 启动配置监听
	e.configWatcher.Start(ctx)

	log.Printf("Executor %s started successfully", e.config.NodeID)

	return nil
}

// Stop 停止执行器
func (e *Executor) Stop(ctx context.Context) error {
	log.Printf("Stopping executor %s", e.config.NodeID)

	// 1. 停止配置监听
	if e.configWatcher != nil {
		e.configWatcher.Stop()
	}

	// 2. 停止触发器
	if e.triggerManager != nil {
		e.triggerManager.StopAll()
	}

	// 3. 停止心跳
	if e.heartbeat != nil {
		e.heartbeat.Stop()
	}

	// 4. 关闭etcd连接
	if e.etcdClient != nil {
		e.etcdClient.Close()
	}

	return nil
}

// connectEtcd 连接etcd
func (e *Executor) connectEtcd() error {
	cfg := clientv3.Config{
		Endpoints:   []string{e.config.EtcdEndpoints},
		Username:    e.config.EtcdUsername,
		Password:    e.config.EtcdPassword,
		DialTimeout: 5 * time.Second,
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		return err
	}

	e.etcdClient = client
	return nil
}

// loadInitialConfig 加载初始配置
func (e *Executor) loadInitialConfig(ctx context.Context) error {
	// 从etcd加载当前配置
	// 1. 加载触发器
	// 2. 加载流程
	// 3. 加载资源
	// 4. 加载工具

	log.Printf("Loading initial configuration for cluster %s", e.config.ClusterID)

	// TODO: 实现配置加载逻辑
	// 从 /clusters/{cluster-id}/triggers/* 加载触发器
	// 从 /clusters/{cluster-id}/flows/* 加载流程

	return nil
}
