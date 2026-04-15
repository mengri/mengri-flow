# TASK-001 工作空间成员权限访问

**优先级**: 高
**模块**: workspace
**涉及层**: App Service / Domain

---

## 背景

`GetWorkspace`、`UpdateWorkspace`、`DeleteWorkspace` 等接口的权限检查只允许 Owner 通过，非 Owner 一律返回 `ErrForbidden`。然而 `AddMember` 功能已实现，工作空间成员实际无法访问自己被邀请加入的工作空间，成员管理体系形同虚设。

## 问题定位

```
internal/app/service/workspace_service.go:124-129
// TODO: 实现成员权限检查
if workspace.OwnerID != accountID {
    slog.Warn(...)
    return nil, domainErr.ErrForbidden  // 所有非 Owner 均被拒绝
}
```

## 需要修改的方法

| 方法 | 当前行为 | 期望行为 |
|------|----------|----------|
| `GetWorkspace` | 仅 Owner 可访问 | Owner 或任意成员可访问 |
| `UpdateWorkspace` | 仅 Owner 可修改 | Owner 或 Admin 成员可修改 |
| `DeleteWorkspace` | 仅 Owner 可删除 | 仅 Owner 可删除（正确，保持） |
| `ListWorkspaces` | 仅返回自己 Own 的 | 同时返回作为成员加入的工作空间 |

## 实现方案

### 1. 提取公共权限检查辅助方法

```go
// checkWorkspaceReadAccess 检查是否为 Owner 或成员
func (s *workspaceServiceImpl) checkWorkspaceReadAccess(
    ctx context.Context, workspace *entity.Workspace, accountID string,
) error {
    if workspace.OwnerID == accountID {
        return nil
    }
    _, err := s.memberRepo.FindByWorkspaceIDAndAccountID(ctx, workspace.ID, accountID)
    if err != nil {
        return domainErr.ErrForbidden
    }
    return nil
}

// checkWorkspaceWriteAccess 检查是否为 Owner 或 Admin 成员
func (s *workspaceServiceImpl) checkWorkspaceWriteAccess(
    ctx context.Context, workspace *entity.Workspace, accountID string,
) error {
    if workspace.OwnerID == accountID {
        return nil
    }
    member, err := s.memberRepo.FindByWorkspaceIDAndAccountID(ctx, workspace.ID, accountID)
    if err != nil || member.Role != entity.MemberRoleAdmin {
        return domainErr.ErrForbidden
    }
    return nil
}
```

### 2. 修改 ListWorkspaces

`WorkspaceRepository` 接口增加 `ListByMember` 方法，返回账号作为成员加入的工作空间（包含 Own 的），使结果集合并去重后返回正确的 total。

```go
// domain/repository/workspace_repository.go
ListByMemberOrOwner(ctx context.Context, accountID string, offset, limit int) ([]*entity.Workspace, int64, error)
```

## 验收标准

- [ ] 成员账号能正常调用 `GET /workspaces/:id`
- [ ] 成员账号能正常调用 `GET /workspaces`（含已加入的工作空间）
- [ ] Admin 成员能调用 `PUT /workspaces/:id`
- [ ] 普通成员调用 `PUT/DELETE /workspaces/:id` 返回 403
- [ ] 所有修改附带单元测试

## 相关文件

- `internal/app/service/workspace_service.go`
- `internal/domain/repository/workspace_repository.go`
- `internal/infra/persistence/mysql/workspace_repository/repository.go`
