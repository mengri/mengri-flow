# 任务 18: Executor执行器

## 任务概述
实现Executor角色，包括配置监听、触发器管理、流程执行引擎、状态上报、心跳检测。

## 上下文依赖
- 任务 01: 基础框架
- 任务 02: 数据库实体
- 任务 03: 插件框架核心
- 任务 07-09: 触发器插件
- 任务 12: 流程编排模块

## 涉及文件
- `cmd/executor/main.go` - Executor入口
- `internal/executor/config_watcher.go` - etcd配置监听
- `internal/executor/trigger_manager.go` - 触发器管理器
- `internal/executor/flow_engine.go` - 流程执行引擎
- `internal/executor/heartbeat.go` - 心跳上报

## 详细步骤

### 18.1 Executor入口
**文件：`cmd/executor/main.go`**
```go
package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    
    "backend/internal/executor"
    "backend/internal/infra/config"
)

func main() {
    // 1. 解析命令行参数
    var (
        etcdEndpoints = flag.String("etcd-endpoints", "localhost:2379", "etcd endpoints")
        etcdUsername  = flag.String("etcd-username", "", "etcd username")
        etcdPassword  = flag.String("etcd-password", "", "etcd password")
        clusterID     = flag.String("cluster-id", "", "cluster ID")
        nodeID        = flag.String("node-id", "", "executor node ID (auto-generated if empty)")
        logLevel      = flag.String("log-level", "info", "log level")
    )
    flag.Parse()
    
    if *clusterID == "" {
        log.Fatal("cluster-id is required")
    }
    
    // 2. 生成Node ID（如果不指定）
    if *nodeID == "" {
        *nodeID = generateNodeID()
    }
    
    // 3. 加载配置
    cfg := &config.ExecutorConfig{
        EtcdEndpoints: *etcdEndpoints,
        EtcdUsername:  *etcdUsername,
        EtcdPassword:  *etcdPassword,
        ClusterID:     *clusterID,
        NodeID:        *nodeID,
        LogLevel:      *logLevel,
    }
    
    // 4. 初始化Executor
    executor := executor.NewExecutor(cfg)
    
    // 5. 启动
    ctx := context.Background()
    if err := executor.Start(ctx); err != nil {
        log.Fatal("Failed to start executor:", err)
    }
    
    // 6. 等待信号
    waitForShutdownSignal()
    
    // 7. 优雅关闭
    if err := executor.Stop(ctx); err != nil {
        log.Fatal("Failed to stop executor:", err)
    }
}

func generateNodeID() string {
    hostname, _ := os.Hostname()
    return fmt.Sprintf("%s-%d", hostname, time.Now().Unix())
}
```

### 18.2 Executor核心结构
**文件：`internal/executor/executor.go`**
```go
package executor

type Executor struct {
    config          *config.ExecutorConfig
    etcdClient      *clientv3.Client
    triggerManager  *TriggerManager
    flowEngine      *FlowEngine
    heartbeat       *Heartbeat
    configWatcher   *ConfigWatcher
}

func NewExecutor(cfg *config.ExecutorConfig) *Executor {
    return &Executor{
        config: cfg,
    }
}

func (e *Executor) Start(ctx context.Context) error {
    log.Printf("Starting executor %s for cluster %s", e.config.NodeID, e.config.ClusterID)
    
    // 1. 连接etcd
    if err := e.connectEtcd(); err != nil {
        return err
    }
    
    // 2. 初始化组件
    e.flowEngine = NewFlowEngine(e.config.NodeID)
    e.triggerManager = NewTriggerManager(e.pluginRegistry, e.flowEngine, e.config.NodeID)
    e.heartbeat = NewHeartbeat(e.etcdClient, e.config.ClusterID, e.config.NodeID)
    e.configWatcher = NewConfigWatcher(e.etcdClient, e.config.ClusterID, e.triggerManager)
    
    // 3. 启动心跳
    e.heartbeat.Start(ctx)
    
    // 4. 加载当前配置
    if err := e.loadInitialConfig(ctx); err != nil {
        return err
    }
    
    // 5. 启动配置监听
    e.configWatcher.Start(ctx)
    
    log.Printf("Executor %s started successfully", e.config.NodeID)
    
    return nil
}

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
```

### 18.3 配置监听
**文件：`internal/executor/config_watcher.go`**
```go
type ConfigWatcher struct {
    etcdClient     *clientv3.Client
    clusterID      string
    triggerManager *TriggerManager
    watchChan      clientv3.WatchChan
}

func NewConfigWatcher(client *clientv3.Client, clusterID string, triggerManager *TriggerManager) *ConfigWatcher {
    return &ConfigWatcher{
        etcdClient:     client,
        clusterID:      clusterID,
        triggerManager: triggerManager,
    }
}

func (w *ConfigWatcher) Start(ctx context.Context) {
    // 监听前缀
    prefix := fmt.Sprintf("/clusters/%s/", w.clusterID)
    
    // 创建watch
    w.watchChan = w.etcdClient.Watch(ctx, prefix, clientv3.WithPrefix())
    
    // 处理变更
    go func() {
        for resp := range w.watchChan {
            for _, event := range resp.Events {
                w.handleEtcdEvent(event)
            }
        }
    }()
    
    log.Printf("Config watcher started for %s", prefix)
}

func (w *ConfigWatcher) handleEtcdEvent(event *clientv3.Event) {
    key := string(event.Kv.Key)
    
    // 解析key格式: /clusters/{cluster-id}/{type}/{id}
    parts := strings.Split(key, "/")
    if len(parts) < 5 {
        return
    }
    
    resourceType := parts[3]
    resourceID := parts[4]
    
    switch event.Type {
    case clientv3.EventTypePut:
        // 创建或更新
        switch resourceType {
        case "triggers":
            var trigger entity.Trigger
            json.Unmarshal(event.Kv.Value, &trigger)
            w.triggerManager.AddOrUpdateTrigger(&trigger)
            
        case "flows":
            // 更新流程缓存
            
        case "resources":
            // 更新资源缓存
            
        case "tools":
            // 更新工具缓存
        }
        
    case clientv3.EventTypeDelete:
        // 删除
        switch resourceType {
        case "triggers":
            w.triggerManager.RemoveTrigger(resourceID)
        }
    }
}
```

### 18.4 触发器管理器
**文件：`internal/executor/trigger_manager.go`**
```go
type TriggerManager struct {
    pluginRegistry *plugin.Registry
    flowEngine     *FlowEngine
    triggers       map[string]plugin.TriggerPlugin
    handlers       map[string]plugin.TriggerHandler
    mu             sync.RWMutex
}

func (m *TriggerManager) AddOrUpdateTrigger(trigger *entity.Trigger) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    // 1. 如果已存在，先停止旧的
    if oldPlugin, exists := m.triggers[trigger.ID.String()]; exists {
        oldPlugin.Stop()
    }
    
    // 2. 获取触发器插件
    plugin, err := m.pluginRegistry.GetTrigger(trigger.Type)
    if err != nil {
        return err
    }
    
    // 3. 创建handler
    handler := m.createTriggerHandler(trigger)
    
    // 4. 启动触发器
    config := trigger.Config
    config["triggerId"] = trigger.ID.String()
    
    ctx := context.Background()
    if err := plugin.Start(ctx, config, handler); err != nil {
        return err
    }
    
    // 5. 保存
    m.triggers[trigger.ID.String()] = plugin
    m.handlers[trigger.ID.String()] = handler
    
    log.Printf("Trigger %s started", trigger.ID)
    
    return nil
}

func (m *TriggerManager) createTriggerHandler(trigger *entity.Trigger) plugin.TriggerHandler {
    return func(ctx context.Context, input map[string]interface{}) (*plugin.TriggerResult, error) {
        // 1. 输入映射
        flowInput := m.applyInputMapping(input, trigger.InputMapping)
        
        // 2. 查询流程
        flow, err := m.flowEngine.GetFlow(trigger.FlowID.String())
        if err != nil {
            return nil, err
        }
        
        // 3. 执行流程
        output, err := m.flowEngine.ExecuteFlow(ctx, flow, flowInput)
        if err != nil {
            // 错误处理
            errorOutput := m.applyErrorHandling(err, trigger.ErrorHandling)
            return &plugin.TriggerResult{
                Success: false,
                Error:   err.Error(),
                Data:    errorOutput,
            }, nil
        }
        
        // 4. 输出映射
        response := m.applyOutputMapping(output, trigger.OutputMapping)
        
        return &plugin.TriggerResult{
            Success: true,
            Data:    response,
        }, nil
    }
}
```

### 18.5 流程执行引擎
**文件：`internal/executor/flow_engine.go`**
```go
type FlowEngine struct {
    nodeID        string
    flowCache     map[string]*entity.Flow
    toolCache     map[string]*entity.Tool
    resourceCache map[string]*entity.Resource
    pluginRegistry *plugin.Registry
}

func (e *FlowEngine) ExecuteFlow(ctx context.Context, flow *entity.Flow, input map[string]interface{}) (map[string]interface{}, error) {
    // 1. 构建执行计划
    plan, err := e.buildExecutionPlan(flow.CanvasData)
    if err != nil {
        return nil, err
    }
    
    // 2. 执行节点
    contextData := make(map[string]interface{})
    contextData["start"] = input
    
    for _, node := range plan {
        output, err := e.executeNode(ctx, node, contextData)
        if err != nil {
            return nil, err
        }
        contextData[node.ID] = output
    }
    
    // 3. 返回结果
    return contextData["end"].(map[string]interface{}), nil
}

func (e *FlowEngine) executeNode(ctx context.Context, node *Node, contextData map[string]interface{}) (map[string]interface{}, error) {
    switch node.Type {
    case "tool":
        return e.executeToolNode(ctx, node, contextData)
    default:
        // start/end节点
        return contextData[node.ID].(map[string]interface{}), nil
    }
}

func (e *FlowEngine) executeToolNode(ctx context.Context, node *Node, contextData map[string]interface{}) (map[string]interface{}, error) {
    // 1. 获取工具
    tool := e.toolCache[node.ToolID]
    
    // 2. 获取资源
    resource := e.resourceCache[tool.ResourceID.String()]
    
    // 3. 获取插件
    plugin, err := e.pluginRegistry.GetResource(resource.Type)
    if err != nil {
        return nil, err
    }
    
    // 4. 参数映射
    input := e.mapInput(node, contextData)
    
    // 5. 执行
    result, err := plugin.ExecuteTool(ctx, resource.Config, tool.Config, input)
    if err != nil {
        return nil, err
    }
    
    return result.Data.(map[string]interface{}), nil
}
```

### 18.6 心跳上报
**文件：`internal/executor/heartbeat.go`**
```go
type Heartbeat struct {
    etcdClient  *clientv3.Client
    clusterID   string
    nodeID      string
    interval    time.Duration
    ttl         time.Duration
    stopChan    chan struct{}
}

func (h *Heartbeat) Start(ctx context.Context) {
    // 创建租约
    leaseResp, err := h.etcdClient.Grant(ctx, int64(h.ttl.Seconds()))
    if err != nil {
        log.Printf("Failed to create lease: %v", err)
        return
    }
    
    leaseID := leaseResp.ID
    
    // 启动心跳goroutine
    go func() {
        ticker := time.NewTicker(h.interval)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                // 续约租约
                _, err := h.etcdClient.KeepAliveOnce(ctx, leaseID)
                if err != nil {
                    log.Printf("Failed to keep alive lease: %v", err)
                    
                    // 重新创建租约
                    leaseResp, _ := h.etcdClient.Grant(ctx, int64(h.ttl.Seconds()))
                    leaseID = leaseResp.ID
                }
                
                // 更新执行器状态
                h.updateExecutorStatus(ctx, leaseID)
                
            case <-h.stopChan:
                return
            }
        }
    }()
}

func (h *Heartbeat) updateExecutorStatus(ctx context.Context, leaseID clientv3.LeaseID) {
    // 构建状态信息
    status := &ExecutorStatus{
        NodeID:        h.nodeID,
        IP:            getLocalIP(),
        Hostname:      getHostname(),
        Version:       getVersion(),
        Status:        "active",
        CPUUsage:      getCPUUsage(),
        MemoryUsage:   getMemoryUsage(),
        RunningFlows:  getRunningFlows(),
        LastHeartbeat: time.Now(),
    }
    
    // 序列化
    data, _ := json.Marshal(status)
    
    // 写入etcd
    key := fmt.Sprintf("/clusters/%s/executors/%s", h.clusterID, h.nodeID)
    
    _, err := h.etcdClient.Put(ctx, key, string(data), clientv3.WithLease(leaseID))
    if err != nil {
        log.Printf("Failed to update executor status: %v", err)
    }
}
```

### 18.7 配置示例
**Executor启动命令：**
```bash
./executor \
  --etcd-endpoints=etcd-1:2379,etcd-2:2379,etcd-3:2379 \
  --cluster-id=cluster-prod-001 \
  --node-id=executor-node-1 \
  --log-level=info

# Docker
Docker run -d \
  --name executor-prod-1 \
  -e ETCD_ENDPOINTS=etcd-1:2379 \
  mengri-flow/executor:latest \
  --cluster-id=cluster-prod-001
```

## 验收标准
- [ ] Executor可成功连接etcd
- [ ] 配置监听正确响应变更
- [ ] 触发器启动和停止正常
- [ ] 流程执行正确
- [ ] 心跳上报正常（5秒间隔）
- [ ] 多实例部署时通过分布式锁协调
- [ ] 优雅关闭清理资源

## 参考文档
- `docs/architecture-design.md` - Executor设计
- `docs/PRD.md` - 第3章系统架构概览

## 预估工时
5-6 天
