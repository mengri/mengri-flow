# TASK-003 激活链接安全性改进

**优先级**: 高
**模块**: auth
**涉及层**: App Service / Infra

---

## 背景

管理员创建账号时，系统发送激活邮件，激活链接中直接包含原始 token（`?token=<rawToken>`）。原始 token 若被邮件拦截或转发，攻击者可直接使用。当前实现已在数据库中存储 SHA-256 哈希，但传输层仍然暴露了原始值。

## 问题定位

```
internal/app/service/account_admin_service.go:66
activationLink := fmt.Sprintf("%s?token=%s", s.cfg.Email.Activation.BaseURL, rawToken)
```

## 实现方案

### 方案 A：使用 token ID 替代原始 token（推荐）

1. `ActivationToken` 实体增加 `TokenID` 字段（UUID），作为公开标识
2. 激活链接改为 `?token_id=<tokenID>`
3. 激活接口先通过 `tokenID` 查出记录，再由用户输入密码 + 提交 token 进行验证

```
激活链接: https://app.example.com/activate?tid=abc-123
激活页面: 用户输入新密码 → POST /api/v1/auth/activate {tid, password, token}
后端:    FindByTokenID(tid) → 比较 hash(token) == storedHash → 激活账号
```

### 方案 B：缩短 token 有效期 + 一次性使用

当前 token 有效期默认 86400 秒（24 小时），可缩短至 3600 秒（1 小时），并确保 `MarkUsed` 在激活成功后立即调用。此方案改动较小，但安全性提升有限。

## 迁移步骤（方案 A）

1. `ActivationToken` 实体和数据库模型增加 `token_id` 列
2. 新增 `FindByTokenID(ctx, tokenID)` 仓储方法
3. 修改激活链接生成逻辑
4. 修改激活 API 接口：接受 `token_id` + `token` 而非仅 `token`
5. 前端激活页面适配新接口
6. 数据库迁移：`ALTER TABLE activation_tokens ADD COLUMN token_id VARCHAR(36)`

## 验收标准

- [ ] 激活链接不再包含原始 token 明文
- [ ] 原始 token 仅在用户提交时通过 HTTPS 传输
- [ ] 已发出的旧格式链接在过渡期内仍可使用
- [ ] 新旧格式的自动化测试

## 相关文件

- `internal/domain/entity/activation_token.go`
- `internal/app/service/account_admin_service.go`
- `internal/app/service/auth_service.go`（ActivateAccount 方法）
- `internal/domain/repository/activation_token_repository.go`
- `internal/infra/persistence/mysql/activation_token_repository/`
- `web/src/views/ActivationView.vue`
