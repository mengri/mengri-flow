# TASK-009 Service 层单元测试补充

**优先级**: 中
**模块**: 全局
**涉及层**: App Service

---

## 背景

当前测试仅覆盖 Domain Entity 和 Infra Auth 层，Service 层（业务核心）完全缺乏单元测试。任何 Service 层的逻辑变更都无法自动验证正确性。

## 当前测试覆盖情况

| 层级 | 已有测试 | 状态 |
|------|----------|------|
| Domain Entity | account, workspace, identity, activation_token | OK |
| Domain ValueObject | email, password, phone | OK |
| Infra Auth | jwt, password | OK |
| **App Service** | **全部缺失** | **缺失** |
| **HTTP Handler** | **全部缺失** | **缺失** |
| **Repository** | **全部缺失** | **缺失** |

## 实现方案

### 优先级排序（按业务风险）

| 优先级 | 服务 | 理由 |
|--------|------|------|
| P0 | AuthService | 认证是安全核心，逻辑最复杂 |
| P0 | WorkspaceService | 权限检查逻辑易出错 |
| P1 | FlowService | 核心业务 |
| P1 | ToolService | 插件相关 |
| P2 | AccountAdminService | 管理员操作 |
| P2 | MeService | 用户自助操作 |
| P2 | TriggerService | 触发器逻辑 |
| P3 | ResourceService | 资源管理 |
| P3 | RunService | 运行管理 |
| P3 | ClusterService | 集群管理 |
| P3 | EnvironmentService | 环境管理 |

### 测试策略

1. **Mock 仓储**：基于接口生成 mock（推荐 `gomock` 或手写 stub）
2. **表格驱动**：每个方法覆盖 正常路径 / 边界条件 / 错误路径
3. **覆盖率目标**：P0 服务 >= 80%，P1 >= 60%

### AuthService 测试用例示例

```
LoginByPassword:
  - 正常登录
  - 账号不存在
  - 密码错误
  - 账号未激活
  - 账号已锁定
  - 连续失败锁定

LoginBySMS:
  - 正常登录
  - 验证码错误
  - 验证码过期
  - 频率限制

RefreshToken:
  - 正常刷新
  - Refresh Token 已过期
  - Session 不存在
  - 使用 Access Token 刷新（应拒绝）

ActivateAccount:
  - 正常激活
  - Token 过期
  - Token 已使用
  - 重复激活
```

## 验收标准

- [ ] P0 服务（Auth, Workspace）单元测试覆盖率 >= 80%
- [ ] P1 服务（Flow, Tool）单元测试覆盖率 >= 60%
- [ ] CI 中运行 `go test ./internal/app/service/...` 不报错
- [ ] Mock 对象基于接口生成，不依赖真实数据库

## 相关文件

- `internal/app/service/*_service.go`（被测文件）
- `internal/app/service/*_test.go`（新增测试文件）
- `internal/domain/repository/`（mock 目标接口）
