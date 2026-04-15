# TASK-008 响应格式统一化

**优先级**: 低
**模块**: 全局
**涉及层**: HTTP Handler / pkg/response

---

## 背景

Handler 层使用了多种不同的响应方法，且参数格式不统一，增加维护成本和出错概率。

## 问题定位

```
// 同一文件中出现多种写法：
response.Success(c, workspace)                                    // 统一格式
response.OK(c, resp)                                             // 另一个方法
response.Success(c, gin.H{"message": "workspace deleted successfully"})  // 混合 gin.H
```

## 问题分析

1. `response.Success` 和 `response.OK` 功能重复，语义不清晰
2. `gin.H` 作为 data 传入绕过了类型约束
3. 删除等无返回值的操作缺乏统一的空响应约定

## 实现方案

### 1. 统一响应方法

保留以下方法，移除冗余：

| 方法 | 场景 | HTTP 状态码 |
|------|------|------------|
| `response.OK(c, data)` | 查询成功，返回数据 | 200 |
| `response.Created(c, data)` | 创建成功 | 201 |
| `response.NoContent(c)` | 删除/更新成功，无返回体 | 204 |
| `response.BadRequest(c, msg)` | 参数错误 | 400 |
| `response.Unauthorized(c, msg)` | 未认证 | 401 |
| `response.Forbidden(c, msg)` | 无权限 | 403 |
| `response.NotFound(c, msg)` | 资源不存在 | 404 |
| `response.Conflict(c, msg)` | 冲突 | 409 |
| `response.InternalError(c, msg)` | 内部错误 | 500 |

### 2. 移除 `gin.H` 直接传参

删除操作的 Handler 改为：
```go
// Before
response.Success(c, gin.H{"message": "workspace deleted successfully"})

// After
response.NoContent(c)
```

### 3. 搜索替换

全量搜索 `response.Success` 和 `response.OK` 的使用，按语义替换为对应方法。

## 验收标准

- [ ] `pkg/response` 只保留上述 9 个方法
- [ ] 所有 Handler 不再使用 `gin.H` 传入 data
- [ ] 删除操作返回 204 No Content
- [ ] 创建操作返回 201 Created
- [ ] 全局 grep 无残留的旧方法调用

## 相关文件

- `pkg/response/response.go`
- `internal/ports/http/handler/*.go`（所有 Handler）
