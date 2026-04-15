# TASK-007 日志敏感信息脱敏

**优先级**: 中
**模块**: 全局
**涉及层**: App Service

---

## 背景

多处日志记录直接输出了用户手机号、邮箱等 PII（个人身份信息），不符合数据保护合规要求（GDPR/个人信息保护法）。

## 问题定位

```
internal/app/service/auth_service.go:413
slog.Error("failed to send sms", "phone", req.Phone, "error", err)
// 完整手机号被记录在日志中

internal/app/service/auth_service.go (多处)
slog.Info("...", "accountId", accountID, "email", account.Email.String(), ...)
// 完整邮箱被记录
```

## 脱敏规则

| 字段 | 当前 | 期望 |
|------|------|------|
| 手机号 | `13800138000` | `138****8000` |
| 邮箱 | `user@example.com` | `us***@example.com` |
| IP 地址 | `192.168.1.100` | `192.168.*.*`（按需） |

> 注：`accountId`（UUID）和 `workspaceId`（UUID）属于系统内部标识，不属于 PII，可保持原样。

## 实现方案

### 1. 统一脱敏工具函数

项目中 `auth_service.go` 已有 `maskEmail` 和 `maskPhone` 私有方法，应提取到公共包：

```go
// pkg/mask/mask.go
package mask

func Email(email string) string { ... }  // u***@example.com
func Phone(phone string) string { ... }  // 138****8000
func IP(ip string) string string { ... } // 192.168.*.*
```

### 2. 全局 grep 替换

扫描所有 `slog.*` 调用中使用 `"phone"` 或 `"email"` key 的地方，替换为脱敏版本。

### 3. 可选：slog Handler 自动脱敏

实现自定义 `slog.Handler`，对指定 key（phone, email, ip）自动脱敏，无需每处手动调用：

```go
type SanitizeHandler struct {
    slog.Handler
    sensitiveKeys map[string]bool
}
```

## 验收标准

- [ ] 日志中不再出现完整手机号和邮箱
- [ ] `pkg/mask` 包提供 `Email`、`Phone`、`IP` 函数
- [ ] 全部 `slog` 调用点已替换
- [ ] grep 验证无遗漏：`grep -r '"phone".*req\.' internal/` 返回空

## 相关文件

- `internal/app/service/auth_service.go`（多处 slog 调用）
- `internal/app/service/account_admin_service.go`
- `internal/app/service/me_service.go`
- `pkg/mask/`（新增）
