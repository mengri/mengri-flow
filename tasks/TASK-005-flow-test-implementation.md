# TASK-005 FlowService.TestFlow 真实实现

**优先级**: 中
**模块**: flow
**涉及层**: App Service / Executor

---

## 背景

`FlowService.TestFlow` 是流程测试接口，当前直接返回 `testSuccess = true`，未执行任何实际测试。用户可能误以为流程已通过验证。

## 问题定位

```
internal/app/service/flow_service.go:233-247
// TODO: 实现流程测试逻辑
// 1. 解析流程配置
// 2. 构建执行上下文
// 3. 模拟执行流程节点
// 4. 验证输出结果
testSuccess := true
```

## 实现方案

### Phase 1：静态校验（快速实现）

在不启动真实执行引擎的情况下，对流程定义做静态验证：

1. **结构校验**：节点 ID 唯一、入口节点存在
2. **连接校验**：边的 source/target 节点存在、无孤立节点（除入口/出口）
3. **类型校验**：每个节点的 `type` 与已注册插件匹配、必要参数齐全
4. **循环检测**：有向图无环（或仅在允许的循环区域内）

```go
type FlowValidationResult struct {
    Valid    bool              `json:"valid"`
    Errors   []ValidationError `json:"errors,omitempty"`
    Warnings []ValidationError `json:"warnings,omitempty"`
}

type ValidationError struct {
    NodeID  string `json:"nodeId,omitempty"`
    EdgeID  string `json:"edgeId,omitempty"`
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

### Phase 2：模拟执行（后续迭代）

1. 使用 Mock 插件替代真实资源插件
2. 按拓扑序逐步执行节点
3. 收集每步输入/输出快照
4. 返回完整的执行追踪

## 验收标准

- [ ] Phase 1：空流程、孤立节点、无效类型、循环依赖均能正确报错
- [ ] Phase 1：合法流程返回 `valid: true`
- [ ] 测试接口响应包含具体错误信息，而非简单的布尔值
- [ ] 单元测试覆盖常见校验场景

## 相关文件

- `internal/app/service/flow_service.go`
- `internal/domain/entity/flow.go`
- `internal/infra/plugin/registry.go`
