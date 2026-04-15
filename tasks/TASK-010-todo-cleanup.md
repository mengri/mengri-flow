# TASK-010 未完成功能 TODO 清理

**优先级**: 低
**模块**: 全局
**涉及层**: 多层

---

## 背景

代码中存在多处 `TODO` 标记，部分是功能缺失，部分是临时方案。需要逐项评估并处理，避免 TODO 积压成为技术债务。

## 当前 TODO 清单

| 位置 | 内容 | 类型 | 建议 |
|------|------|------|------|
| `workspace_service.go:125` | 成员权限检查 | 功能缺失 | -> TASK-001 |
| `workspace_service.go:219` | 删除前资源检查 | 数据完整性 | -> TASK-002 |
| `flow_service.go:233` | 流程测试逻辑 | 功能缺失 | -> TASK-005 |
| `flow_service.go:345` | 版本历史获取 | 功能缺失 | 本任务处理 |
| `me_handler.go:76` | 绑定手机号 | 功能缺失 | 本任务处理 |
| `executor/executor.go` | loadInitialConfig | 节点启动 | 本任务处理 |

## 实现方案

### 1. 已有任务卡片的 TODO

标注为关联已有任务卡片，不在本任务中重复实现：

```go
// workspace_service.go:125
// TODO(workspace-member-access): 实现成员权限检查 → 见 TASK-001
```

### 2. Flow 版本历史获取

```go
// flow_service.go:345
// TODO: 从版本管理系统中获取流程的所有版本
```

**方案**：
- `FlowRepository` 增加 `ListVersionsByFlowID(ctx, flowID, offset, limit)` 方法
- 数据库增加 `flow_versions` 表（flow_id, version, config_snapshot, created_by, created_at）
- `Flow.Publish()` 时自动创建版本快照
- 新增 `GET /flows/:id/versions` 接口

### 3. 绑定手机号

```go
// me_handler.go:76
// TODO: 实现绑定手机号
```

**方案**：`MeService` 中已有 `BindPhone` 方法实现，但 Handler 路由被注释。需要：
- 取消 `router.go` 中手机绑定路由的注释
- 前端添加手机号绑定页面
- 端到端测试验证

### 4. Executor loadInitialConfig

```go
// executor/executor.go
// TODO: loadInitialConfig
```

**方案**：
- 从 etcd `/config/{clusterID}/triggers` 路径读取初始触发器配置
- 逐个调用 `TriggerManager.StartTrigger`
- 记录启动日志

## 验收标准

- [ ] 所有 TODO 标注了关联任务卡片编号或本任务的处理方案
- [ ] 手机号绑定路由取消注释并可用
- [ ] `flow_versions` 表和 API 实现
- [ ] Executor 启动时从 etcd 加载初始配置
- [ ] 代码中无孤立的 `TODO` 标记

## 相关文件

- `internal/app/service/workspace_service.go`
- `internal/app/service/flow_service.go`
- `internal/ports/http/handler/me_handler.go`
- `internal/ports/http/router/router.go`
- `internal/executor/executor.go`
