# 任务 08: Timer触发器插件

## 任务概述
实现Timer触发器插件，支持Cron表达式定时触发流程执行，并通过etcd分布式锁确保集群中只有一个实例执行定时任务。

## 上下文依赖
- 任务 03: 插件框架核心
- 任务 02: 数据库实体（Trigger, Run）
- 任务 18: Executor执行器核心
- 需要etcd集群

## 涉及文件
- `plugins/trigger/timer/plugin.go` - Timer触发器插件
- `plugins/trigger/timer/plugin_test.go` - 单元测试
- `internal/infra/distributed/lock.go` - 分布式锁实现

## 详细步骤

### 8.1 插件基本结构
**文件：`plugins/trigger/timer/plugin.go`**
```go
//go:build timer

package timer

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    cron "github.com/robfig/cron/v3"
    clientv3 "go.etcd.io/etcd/client/v3"
    concurrency "go.etcd.io/etcd/client/v3/concurrency"
    "backend/internal/infra/plugin"
)

func init() {
    registry := plugin.GlobalRegistry()
    registry.RegisterTrigger(&TimerTriggerPlugin{})
}

type TimerTriggerPlugin struct {
    mu       sync.RWMutex
    crons    map[string]*cron.Cron          // triggerID -> cron实例
    configs  map[string]map[string]interface{}
    handlers map[string]plugin.TriggerHandler
    etcdClient *clientv3.Client
}
```
- [ ] 创建插件文件，添加 `//go:build timer` 构建标签
- [ ] 使用 robfig/cron/v3 库
- [ ] 集成etcd客户端

### 8.2 实现插件元数据
```go
func (p *TimerTriggerPlugin) PluginMeta() plugin.PluginMeta {
    return plugin.PluginMeta{
        Name:        "timer",
        Type:        plugin.PluginTypeTrigger,
        Version:     "1.0.0",
        Description: "定时任务触发器插件（Cron表达式），支持分布式锁",
        Author:      "Platform Team",
        BuildTag:    "timer",
    }
}
```

### 8.3 实现配置Schema
```go
func (p *TimerTriggerPlugin) ConfigSchema() plugin.JSONSchema {
    return plugin.JSONSchema{
        "type": "object",
        "required": []string{"cronExpression"},
        "properties": map[string]interface{}{
            "cronExpression": map[string]interface{}{
                "type":        "string",
                "title":       "Cron表达式",
                "description": "标准的Cron表达式，如 0 2 * * *（每天凌晨2点）",
                "placeholder": "0 2 * * *",
                "pattern":     "^([\\*0-9,\\-/]+\\s){4,5}[\\*0-9,\\-/]+$",
            },
            "timezone": map[string]interface{}{
                "type":        "string",
                "title":       "时区",
                "description": "Cron表达式使用的时区",
                "default":     "Asia/Shanghai",
                "enum":        []string{"Asia/Shanghai", "UTC", "America/New_York"},
            },
            "distributedLock": map[string]interface{}{
                "type":        "boolean",
                "title":       "启用分布式锁",
                "description": "确保集群中只有一个实例执行定时任务",
                "default":     true,
            },
            "lockKey": map[string]interface{}{
                "type":        "string",
                "title":       "分布式锁键",
                "description": "etcd中的锁路径，如 /locks/timer-{trigger-id}",
                "condition":   map[string]interface{}{"distributedLock": true},
            },
            "lockTTL": map[string]interface{}{
                "type":        "integer",
                "title":       "锁TTL（秒）",
                "description": "分布式锁的过期时间，应大于任务执行时间",
                "default":     60,
                "minimum":     10,
                "maximum":     3600,
                "condition":   map[string]interface{}{"distributedLock": true},
            },
        },
    }
}
```
- [ ] 包含Cron表达式配置
- [ ] 包含时区配置
- [ ] 包含分布式锁配置（默认启用）

### 8.4 实现输入输出Schema
```go
func (p *TimerTriggerPlugin) InputSchema() plugin.JSONSchema {
    return plugin.JSONSchema{
        "type": "object",
        "properties": map[string]interface{}{
            "triggerTime": map[string]interface{}{
                "type":        "string",
                "format":      "date-time",
                "description": "触发时间（ISO8601格式）",
            },
            "executionId": map[string]interface{}{
                "type":        "string",
                "description": "执行唯一标识",
            },
            "triggerId": map[string]interface{}{
                "type":        "string",
                "description": "触发器ID",
            },
        },
    }
}

func (p *TimerTriggerPlugin) OutputSchema() plugin.JSONSchema {
    return nil  // Timer触发器无需响应输出
}
```
- [ ] 输入包含触发时间和执行ID
- [ ] 输出为nil（异步触发器）

### 8.5 实现Start方法
```go
func (p *TimerTriggerPlugin) Start(
    ctx context.Context,
    config map[string]interface{},
    handler plugin.TriggerHandler,
) error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    triggerID := config["triggerId"].(string)
    
    // 初始化映射
    if p.crontabs == nil {
        p.crontabs = make(map[string]*cron.Cron)
        p.configs = make(map[string]map[string]interface{})
        p.handlers = make(map[string]plugin.TriggerHandler)
    }
    
    p.configs[triggerID] = config
    p.handlers[triggerID] = handler
    
    // 创建Cron实例
    timezone := "Asia/Shanghai"
    if tz, ok := config["timezone"].(string); ok {
        timezone = tz
    }
    
    loc, err := time.LoadLocation(timezone)
    if err != nil {
        return fmt.Errorf("invalid timezone: %w", err)
    }
    
    c := cron.New(cron.WithLocation(loc))
    
    // 添加定时任务
    cronExpr := config["cronExpression"].(string)
    
    _, err = c.AddFunc(cronExpr, func() {
        p.executeWithLock(ctx, triggerID)
    })
    if err != nil {
        return fmt.Errorf("invalid cron expression: %w", err)
    }
    
    p.crontabs[triggerID] = c
    
    // 启动Cron调度器
    c.Start()
    
    fmt.Printf("Timer trigger %s started with cron: %s\n", triggerID, cronExpr)
    
    return nil
}
```
- [ ] 创建Cron调度器实例
- [ ] 设置时区
- [ ] 添加Cron任务
- [ ] 启动调度器

### 8.6 分布式锁实现
```go
// executeWithLock 使用分布式锁执行任务
func (p *TimerTriggerPlugin) executeWithLock(ctx context.Context, triggerID string) {
    config := p.configs[triggerID]
    handler := p.handlers[triggerID]
    
    // 检查是否启用分布式锁
    useLock := true
    if val, ok := config["distributedLock"].(bool); ok {
        useLock = val
    }
    
    if !useLock {
        // 不使用锁，直接执行
        p.execute(ctx, triggerID, handler, config)
        return
    }
    
    // 获取etcd分布式锁
    lockKey := fmt.Sprintf("/locks/timer/%s", triggerID)
    if val, ok := config["lockKey"].(string); ok && val != "" {
        lockKey = val
    }
    
    // 创建会话
    session, err := concurrency.NewSession(p.etcdClient)
    if err != nil {
        fmt.Printf("Failed to create etcd session: %v\n", err)
        return
    }
    defer session.Close()
    
    // 创建锁
    mutex := concurrency.NewMutex(session, lockKey)
    
    // 尝试获取锁（带超时）
    lockCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    
    if err := mutex.Lock(lockCtx); err != nil {
        fmt.Printf("Failed to acquire lock for trigger %s: %v\n", triggerID, err)
        return
    }
    
    // 在后台goroutine中续期锁（保持持有）
    stopRenewal := make(chan struct{})
    go p.renewLock(ctx, mutex, stopRenewal)
    
    // 执行实际任务
    p.execute(ctx, triggerID, handler, config)
    
    // 停止续期并释放锁
    close(stopRenewal)
    
    if err := mutex.Unlock(ctx); err != nil {
        fmt.Printf("Failed to release lock for trigger %s: %v\n", triggerID, err)
    }
}

// renewLock 在后台续期锁
func (p *TimerTriggerPlugin) renewLock(ctx context.Context, mutex *concurrency.Mutex, stop <-chan struct{}) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-stop:
            return
        case <-ctx.Done():
            return
        case <-ticker.C:
            // 注意：etcd锁的实际续期由会话KeepAlive处理
            // 这里只是保持goroutine活跃
        }
    }
}
```
- [ ] 实现分布式锁获取和释放
- [ ] 锁续期机制（通过etcd会话KeepAlive）
- [ ] 获取锁超时处理

### 8.7 执行任务
```go
// execute 执行定时任务
func (p *TimerTriggerPlugin) execute(
    ctx context.Context,
    triggerID string,
    handler plugin.TriggerHandler,
    config map[string]interface{},
) {
    // 构造输入数据
    executionId := generateExecutionID()
    
    input := map[string]interface{}{
        "triggerTime": time.Now().Format(time.RFC3339),
        "executionId": executionId,
        "triggerId":   triggerID,
    }
    
    // 调用handler触发流程
    result, err := handler(ctx, input)
    if err != nil {
        fmt.Printf("Timer trigger %s execution failed: %v\n", triggerID, err)
        return
    }
    
    if !result.Success {
        fmt.Printf("Timer trigger %s execution failed: %s\n", triggerID, result.Error)
        return
    }
    
    fmt.Printf("Timer trigger %s executed successfully\n", triggerID)
}

// generateExecutionID 生成执行ID
func generateExecutionID() string {
    return fmt.Sprintf("timer-%d", time.Now().UnixNano())
}
```
- [ ] 构造标准输入数据
- [ ] 调用流程处理器
- [ ] 记录执行日志

### 8.8 实现Stop方法
```go
func (p *TimerTriggerPlugin) Stop() error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    for triggerID, cron := range p.crontabs {
        // 停止Cron调度器
        cron.Stop()
        fmt.Printf("Timer trigger %s stopped\n", triggerID)
    }
    
    // 清空映射
    p.crontabs = nil
    p.configs = nil
    p.handlers = nil
    
    return nil
}
```
- [ ] 停止所有Cron调度器
- [ ] 清理资源

### 8.9 单元测试
**文件：`plugins/trigger/timer/plugin_test.go`**
```go
func TestTimerTriggerPlugin_PluginMeta(t *testing.T)
func TestTimerTriggerPlugin_ConfigSchema(t *testing.T)
func TestTimerTriggerPlugin_Start_Stop(t *testing.T)
func TestTimerTriggerPlugin_CronExecution(t *testing.T)
func TestTimerTriggerPlugin_DistributedLock(t *testing.T)
```
- [ ] 测试Cron表达式解析
- [ ] 测试任务调度执行
- [ ] 测试分布式锁（需要mock etcd）
- [ ] 测试并发场景下的单执行保证

### 8.10 集成etcd分布式锁
**文件：`internal/infra/distributed/lock.go`**
```go
package distributed

import (
    "context"
    "time"
    
    clientv3 "go.etcd.io/etcd/client/v3"
    concurrency "go.etcd.io/etcd/client/v3/concurrency"
)

type DistributedLock struct {
    client *clientv3.Client
    session *concurrency.Session
    mutex   *concurrency.Mutex
}

func NewLock(client *clientv3.Client, key string, ttl int) (*DistributedLock, error)
func (l *DistributedLock) Lock(ctx context.Context) error
func (l *DistributedLock) Unlock(ctx context.Context) error
```
- [ ] 封装分布式锁为独立组件
- [ ] 支持可配置的TTL

### 8.11 更新插件配置
**文件：`plugins/plugins.yaml`**
- [ ] 添加 `timer` 到 build_tags

## 验收标准
- [ ] 插件符合插件开发规范
- [ ] 支持标准Cron表达式
- [ ] 支持多时区配置
- [ ] 分布式锁确保单集群单次触发只执行一次
- [ ] 锁续期机制可靠
- [ ] 单元测试覆盖率 > 75%
- [ ] 可成功编译：`go build -tags timer`
- [ ] 实际测试：启动多个Executor实例，只有1个执行定时任务

## 技术难点
- etcd分布式锁的正确使用
- 锁的续期和释放时机
- Cron调度器与分布式锁的协调
- 多实例下的任务分配

## 参考文档
- `docs/plugin-development-guide.md` - 第5章触发器插件开发
- `docs/architecture-design.md` - Timer触发器需etcd分布式锁
- `docs/PRD.md` - 触发器类型配置

## 预估工时
4-5 天（含分布式锁实现和测试）
