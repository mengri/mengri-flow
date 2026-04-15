# TASK-002 删除工作空间前检查关联资源

**优先级**: 高
**模块**: workspace
**涉及层**: App Service

---

## 背景

`DeleteWorkspace` 目前在删除工作空间前未检查是否存在关联的 Flow、Tool、Trigger、Cluster、Environment 等资源，直接执行物理删除。如果数据库存在外键约束则报错，如果没有则产生大量孤立数据。

## 问题定位

```
internal/app/service/workspace_service.go:219-224
// TODO: 检查工作空间是否有关联的资源（流程、工具、触发器、集群等）
slog.Warn("Workspace deletion without resource check - should be implemented", ...)
err = s.workspaceRepo.Delete(ctx, workspaceID)  // 直接删除
```

## 需要修改的服务依赖

`workspaceServiceImpl` 目前没有注入其他资源仓储，需要注入以下依赖来做计数检查：

| 依赖 | 接口 | 所需方法 |
|------|------|----------|
| Flow | `FlowRepository` | `CountByWorkspaceID(ctx, wsID) (int64, error)` |
| Tool | `ToolRepository` | `CountByWorkspaceID(ctx, wsID) (int64, error)` |
| Trigger | `TriggerRepository` | `CountByWorkspaceID(ctx, wsID) (int64, error)` |
| Cluster | `ClusterRepository` | `CountByWorkspaceID(ctx, wsID) (int64, error)` |
| Environment | `EnvironmentRepository` | `CountByWorkspaceID(ctx, wsID) (int64, error)` |

> 各 Repository 接口若尚未定义 `CountByWorkspaceID`，需同步在接口和 GORM 实现中添加。

## 实现方案

### 删除前置检查逻辑

```go
func (s *workspaceServiceImpl) DeleteWorkspace(ctx context.Context, id string, accountID string) error {
    // ... 权限校验 ...

    counts := map[string]int64{}
    if n, _ := s.flowRepo.CountByWorkspaceID(ctx, workspaceID); n > 0 {
        counts["flows"] = n
    }
    if n, _ := s.toolRepo.CountByWorkspaceID(ctx, workspaceID); n > 0 {
        counts["tools"] = n
    }
    // ... 其他资源 ...

    if len(counts) > 0 {
        return fmt.Errorf("%w: workspace contains resources: %v", domainErr.ErrConflict, counts)
    }

    return s.workspaceRepo.Delete(ctx, workspaceID)
}
```

### 错误提示

返回 HTTP 409 Conflict，响应 body 携带具体资源数量，便于前端展示：
```json
{
  "code": 409,
  "msg": "workspace contains resources: flows=3, tools=1",
  "data": null
}
```

## 验收标准

- [ ] 存在关联 Flow/Tool/Trigger/Cluster 的工作空间无法被删除，返回 409
- [ ] 空工作空间可以正常删除
- [ ] 各资源仓储接口增加 `CountByWorkspaceID` 方法
- [ ] 新增方法有对应的 GORM 实现

## 相关文件

- `internal/app/service/workspace_service.go`
- `internal/app/service/workspace_service_iface.go`
- `internal/domain/repository/flow_repository.go`
- `internal/domain/repository/tool_repository.go`
- `internal/domain/repository/trigger_repository.go`
- `internal/domain/repository/cluster_repository.go`
- `internal/domain/repository/environment_repository.go`
- 各对应的 mysql repository 实现
