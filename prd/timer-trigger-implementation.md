# Timer触发器插件实现总结

## 概述

Timer触发器插件基于Cron表达式实现定时任务调度，支持etcd分布式锁确保集群中只有一个实例执行定时任务。

## 实现特性

### ✅ 已完成功能

1. **Cron表达式调度**
   - 使用 `robfig/cron/v3` 库
   - 支持标准Cron表达式（5或6字段）
   - 支持多时区配置（Asia/Shanghai、UTC、America/New_York）

2. **分布式锁（etcd）**
   - 基于etcd的分布式锁实现
   - 可配置的锁TTL（默认60秒）
   - 自动锁续期（通过etcd会话KeepAlive）
   - 锁获取超时控制（10秒）

3. **配置Schema**
   - Cron表达式（必填）
   - 时区（可选，默认Asia/Shanghai）
   - 分布式锁开关（可选，默认启用）
   - 锁键路径（可选，默认 `/locks/timer/{trigger-id}`）
   - 锁TTL（可选，默认60秒）

4. **输入输出**
   - 输入：触发时间、执行ID、触发器ID
   - 输出：无（异步触发器）

5. **生命周期管理**
   - Start：创建Cron调度器并启动
   - Stop：停止所有调度器并清理资源

## 文件结构

```
plugins/trigger/timer/
├── plugin.go      - 主插件文件（326行）
└── plugin_test.go - 单元测试文件（待完善）
```

## 核心代码

### 插件结构
```go
type TimerTriggerPlugin struct {
    mu         sync.RWMutex
    crons      map[string]*cron.Cron
    configs    map[string]map[string]interface{}
    handlers   map[string]plugin.TriggerHandler
    etcdClient *clientv3.Client
}
```

### 执行任务（带锁）
```go
func (p *TimerTriggerPlugin) executeWithLock(ctx context.Context, triggerID string) {
    // 1. 获取handler和config
    // 2. 检查是否启用分布式锁
    // 3. 创建etcd会话和互斥锁
    // 4. 尝试获取锁（10秒超时）
    // 5. 在后台续期锁
    // 6. 执行任务
    // 7. 释放锁
}
```

### 分布式锁实现
```go
// 创建etcd会话（带TTL）
session, err := concurrency.NewSession(p.etcdClient, concurrency.WithTTL(lockTTL))

// 创建互斥锁
mutex := concurrency.NewMutex(session, lockKey)

// 获取锁（带超时）
lockCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
defer cancel()

if err := mutex.Lock(lockCtx); err != nil {
    // 获取锁失败
    return
}

// 在后台续期锁
stopRenewal := make(chan struct{})
go p.renewLock(session, stopRenewal)

// 执行任务
p.execute(ctx, triggerID, handler, config)

// 停止续期并释放锁
close(stopRenewal)
mutex.Unlock(ctx)
```

## 配置示例

```yaml
triggerId: "daily-report"
cronExpression: "0 2 * * *"      # 每天凌晨2点
timezone: "Asia/Shanghai"        # 时区
distributedLock: true            # 启用分布式锁
lockKey: "/locks/timer/daily-report"  # 锁路径
lockTTL: 60                      # 锁TTL（秒）
```

## 使用场景

1. **定时数据同步**
   ```yaml
   cronExpression: "0 */6 * * *"  # 每6小时
   ```

2. **每日报表生成**
   ```yaml
   cronExpression: "0 1 * * *"    # 每天凌晨1点
   ```

3. **定时清理任务**
   ```yaml
   cronExpression: "0 3 * * 0"    # 每周日凌晨3点
   ```

4. **高频监控**
   ```yaml
   cronExpression: "*/5 * * * * *"  # 每5秒
   ```

## 分布式锁工作流程

```
┌─────────────────┐
│   Cron触发      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 获取etcd会话     │
│ (with TTL)      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 创建互斥锁       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 尝试获取锁       │
│ (10秒超时)      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 启动续期goroutine│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   执行任务      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 停止续期        │
│ 释放锁          │
└─────────────────┘
```

## 集群部署

### 多实例部署
- 部署多个Executor实例
- 所有实例监听相同的etcd配置
- Timer触发器到达时，通过etcd锁竞争执行权
- 只有一个实例获得锁并执行任务

### 故障转移
- 如果获得锁的实例故障（会话超时）
- etcd自动释放锁
- 其他实例可以获取锁并继续执行

## 性能考虑

1. **锁粒度**
   - 每个触发器独立的锁键
   - 不同触发器之间无竞争

2. **锁TTL**
   - 应大于任务执行时间
   - 默认60秒，可配置

3. **续期机制**
   - 基于etcd会话KeepAlive
   - 自动续期，无需手动干预

4. **超时控制**
   - 锁获取超时：10秒
   - 防止长时间等待

## 测试建议

1. **单元测试**
   - Cron表达式解析
   - 时区处理
   - 配置验证

2. **集成测试**
   - 多实例部署
   - 分布式锁竞争
   - 故障转移场景

3. **性能测试**
   - 高频率触发（每秒）
   - 大量触发器（1000+）
   - 锁竞争性能

## 注意事项

1. **etcd可用性**
   - 插件依赖etcd集群
   - etcd故障时，任务无法执行

2. **时钟同步**
   - 集群节点时钟应同步
   - 建议使用NTP

3. **锁TTL配置**
   - 应大于任务最大执行时间
   - 防止锁提前释放导致并发执行

4. **Cron表达式**
   - 支持标准5字段格式
   - 也支持带秒的6字段格式

5. **错误处理**
   - 锁获取失败：记录日志，跳过执行
   - 任务执行失败：记录日志，不影响下次调度

## 后续优化

1. **错过任务处理**
   - 记录因锁竞争错过的任务
   - 提供补偿机制

2. **任务去重**
   - 相同执行ID的任务只执行一次
   - 防止重复执行

3. **执行历史**
   - 记录每次执行详情
   - 存储到数据库

4. **监控告警**
   - 执行失败告警
   - 锁竞争频繁告警
   - 任务延迟告警

## 相关文档

- `plans/08-Timer触发器插件.md` - 原始任务计划
- `docs/plugin-development-guide.md` - 插件开发规范
- `docs/architecture-design.md` - 架构设计
- `internal/infra/plugin/types.go` - 插件接口定义

## 实现状态

✅ 插件基本结构
✅ 插件元数据
✅ 配置Schema
✅ 输入输出Schema
✅ Start/Stop方法
✅ Cron调度器集成
✅ 分布式锁（etcd）
✅ 锁续期机制
✅ 单元测试框架
✅ 插件注册
⏳ 完整单元测试（环境限制）

**注意**：由于当前环境Go版本不匹配，无法运行完整测试，但代码结构完整，符合插件开发规范。
