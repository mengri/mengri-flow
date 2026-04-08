# 任务 05: gRPC资源插件

## ⚠️ 项目规范提醒
在开始本任务前，请仔细阅读并遵守以下规范：

### 插件开发铁律
- **必须**在 `init()` 函数中注册插件：`registry.RegisterResource(&gRPCPlugin{})`
- **必须**添加接口编译时检查：`var _ plugin.ResourcePlugin = (*gRPCPlugin)(nil)`
- **必须**使用正确的构建标签：`//go:build grpc`
- **必须**在 `plugins/plugins.yaml` 中注册 `grpc` 构建标签
- **必须**实现所有接口方法（元数据、Schema、连接测试、工具执行）

### gRPC客户端规范
- **必须**使用 `grpc.DialContext(ctx, ...)` 传递上下文
- **必须**设置合理的超时（默认 30 秒，可配置）
- **必须**处理 TLS 配置（启用/跳过验证）
- **必须**优雅关闭连接（`defer conn.Close()`）
- **必须**处理连接错误（DNS、网络、认证）

### Proto文件解析
- **必须**验证 Proto 语法（使用 `protoparse`）
- **必须**处理 import 语句（相对路径、标准库）
- **必须**提取所有 service 和 rpc 方法
- **必须**转换 message 为 JSON Schema
- **必须**处理嵌套 message 和枚举类型

### 动态调用
- **必须**使用 gRPC 反射 API（`grpc.reflection.v1alpha.ServerReflection`）
- **必须**缓存反射结果（避免重复请求）
- **必须**处理未知 message 类型（动态构建）
- **必须**验证输入参数（根据反射的 message 描述）

---

## 任务概述
实现gRPC资源插件，支持连接gRPC服务、执行RPC方法调用和从Proto文件批量导入工具。

## 上下文依赖
- 任务 03: 插件框架核心
- 任务 02: 数据库实体（Tool定义）

## 涉及文件
- `plugins/resource/grpc/plugin.go` - gRPC插件实现
- `plugins/resource/grpc/plugin_test.go` - 单元测试
- `plugins/resource/grpc/proto_parser.go` - Proto文件解析器

## 详细步骤

### 5.1 插件基本结构
**文件：`plugins/resource/grpc/plugin.go`**
```go
//go:build grpc

package grpc

import (
    "context"
    "fmt"
    "time"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "backend/internal/infra/plugin"
)

func init() {
    registry := plugin.GlobalRegistry()
    registry.RegisterResource(&gRPCPlugin{})
}

type gRPCPlugin struct{}
```
- [ ] 创建插件文件，添加 `//go:build grpc` 构建标签
- [ ] 实现 `init()` 注册插件

### 5.2 实现插件元数据
```go
func (p *gRPCPlugin) PluginMeta() plugin.PluginMeta {
    return plugin.PluginMeta{
        Name:        "grpc",
        Type:        plugin.PluginTypeResource,
        Version:     "1.0.0",
        Description: "gRPC服务资源插件",
        Author:      "Platform Team",
        BuildTag:    "grpc",
    }
}
```

### 5.3 实现配置Schema
```go
func (p *gRPCPlugin) ConfigSchema() plugin.JSONSchema {
    return plugin.JSONSchema{
        "type": "object",
        "required": []string{"serverAddress"},
        "properties": map[string]interface{}{
            "serverAddress": map[string]interface{}{
                "type":        "string",
                "title":       "服务器地址",
                "description": "gRPC服务器地址，如 grpc.example.com:50051",
                "pattern":     "^.+:\\d+$",
            },
            "tls": map[string]interface{}{
                "type": "object",
                "title": "TLS配置",
                "properties": map[string]interface{}{
                    "enabled": map[string]interface{}{
                        "type":    "boolean",
                        "title":   "启用TLS",
                        "default": false,
                    },
                    "insecure": map[string]interface{}{
                        "type":    "boolean",
                        "title":   "跳过证书验证",
                        "default": false,
                        "condition": map[string]interface{}{"tls.enabled": true},
                    },
                },
            },
            "timeout": map[string]interface{}{
                "type":        "integer",
                "title":       "超时时间（毫秒）",
                "default":     30000,
                "minimum":     1000,
                "maximum":     300000,
            },
            "protoDefinition": map[string]interface{}{
                "type":        "string",
                "title":       "Proto定义",
                "description": "Protocol Buffer定义（可选，用于导入工具）",
                "format":      "textarea",
            },
        },
    }
}
```
- [ ] 包含服务器地址配置
- [ ] 包含TLS配置（启用/跳过验证）
- [ ] 包含超时配置
- [ ] 包含Proto定义字段（用于导入）

### 5.4 实现连接测试
```go
func (p *gRPCPlugin) TestConnection(ctx context.Context, config map[string]interface{}) error {
    serverAddress, ok := config["serverAddress"].(string)
    if !ok || serverAddress == "" {
        return &plugin.PluginError{
            Type:    "INVALID_CONFIG",
            Message: "serverAddress is required",
        }
    }
    
    opts := []grpc.DialOption{
        grpc.WithBlock(),
        grpc.WithTimeout(5 * time.Second),
    }
    
    // TLS配置
    if tls, ok := config["tls"].(map[string]interface{}); ok {
        if enabled, _ := tls["enabled"].(bool); enabled {
            insecureVal, _ := tls["insecure"].(bool)
            if insecureVal {
                opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
            }
            // 实际生产环境需要加载证书
        }
    } else {
        opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
    }
    
    conn, err := grpc.Dial(serverAddress, opts...)
    if err != nil {
        return &plugin.PluginError{
            Type:    "CONNECTION_FAILED",
            Message: fmt.Sprintf("failed to connect to gRPC server: %v", err),
            Cause:   plugin.ErrConnectionFailed,
        }
    }
    defer conn.Close()
    
    return nil
}
```
- [ ] 实现gRPC连接测试
- [ ] 支持TLS配置
- [ ] 5秒超时限制

### 5.5 实现工具执行
```go
func (p *gRPCPlugin) ExecuteTool(
    ctx context.Context,
    resourceConfig map[string]interface{},
    toolConfig map[string]interface{},
    input interface{},
) (*plugin.ToolResult, error) {
    start := time.Now()
    
    // 1. 建立连接（或从连接池获取）
    conn, err := p.getConnection(resourceConfig)
    if err != nil {
        return nil, err
    }
    
    // 2. 解析工具配置
    service := toolConfig["service"].(string)
    method := toolConfig["method"].(string)
    
    // 3. 动态调用gRPC方法
    // 由于Go是静态类型，这里需要反射或动态客户端
    // 实现方案：使用 grpcurl 类似的反射机制
    result, err := p.invokeRPC(ctx, conn, service, method, input)
    if err != nil {
        return nil, err
    }
    
    duration := time.Since(start)
    
    return &plugin.ToolResult{
        StatusCode: 200,  // gRPC没有HTTP状态码，成功为200
        Data:       result,
        Duration:   duration,
    }, nil
}
```
- [ ] 实现gRPC连接管理（考虑连接池）
- [ ] 实现RPC方法调用
- [ ] 处理输入参数序列化
- [ ] 处理响应反序列化

### 5.6 gRPC反射支持
```go
// invokeRPC 使用gRPC反射动态调用方法
func (p *gRPCPlugin) invokeRPC(
    ctx context.Context,
    conn *grpc.ClientConn,
    service, method string,
    input interface{},
) (interface{}, error) {
    // 1. 获取服务描述（通过反射）
    // 2. 构建请求消息
    // 3. 调用方法
    // 4. 解析响应
    
    // 实际实现需要：
    // - 使用 google.golang.org/grpc/reflection
    // - 或使用 github.com/fullstorydev/grpchan/inprocgrpc
    
    return nil, fmt.Errorf("not fully implemented")
}
```
- [ ] 集成gRPC反射API
- [ ] 动态获取服务和方法描述
- [ ] 动态构建和解析消息

### 5.7 Proto文件解析
**文件：`plugins/resource/grpc/proto_parser.go`**
```go
import (
    "github.com/jhump/protoreflect/desc/protoparse"
    "github.com/jhump/protoreflect/dynamic"
)

// parseProto 解析Proto文件并提取服务方法
type ProtoParser struct{}

func (p *ProtoParser) Parse(data []byte) ([]plugin.ToolDefinition, error) {
    // 1. 解析proto文件
    // 2. 提取所有service
    // 3. 提取每个service的所有rpc方法
    // 4. 转换request/response message为JSON Schema
    
    return nil, nil
}
```
- [ ] 使用 protoreflect 库解析Proto
- [ ] 提取所有服务和方法定义
- [ ] 转换消息结构为JSON Schema
- [ ] 生成ToolDefinition列表

```go
func (p *gRPCPlugin) ExtractTools(ctx context.Context, config map[string]interface{}) ([]plugin.ToolDefinition, error) {
    protoDef, ok := config["protoDefinition"].(string)
    if !ok || protoDef == "" {
        return nil, &plugin.PluginError{
            Type:    "INVALID_CONFIG",
            Message: "protoDefinition is required for extracting tools",
        }
    }
    
    parser := &ProtoParser{}
    return parser.Parse([]byte(protoDef))
}
```
- [ ] 实现ExtractTools方法

### 5.8 单元测试
**文件：`plugins/resource/grpc/plugin_test.go`**
- [ ] 测试插件元数据和Schema
- [ ] 测试连接成功和失败场景
- [ ] 测试工具执行（需要mock gRPC服务器）
- [ ] 测试Proto解析
- [ ] 覆盖率 > 70%

### 5.9 更新插件配置
**文件：`plugins/plugins.yaml`**
- [ ] 添加 `grpc` 到 build_tags

## 验收标准
- [ ] 插件符合插件开发规范
- [ ] 支持gRPC连接和TLS配置
- [ ] 支持通过反射动态调用RPC方法
- [ ] 支持从Proto文件批量导入工具
- [ ] 单元测试覆盖率 > 70%
- [ ] 可成功编译：`go build -tags grpc`

## 技术难点
- gRPC动态调用需要反射支持
- Proto解析和消息构建复杂度高
- 可能需要简化实现（MVP阶段只支持静态生成的客户端）

## 参考文档
- `docs/plugin-development-guide.md` - 资源插件规范
- `docs/PRD.md` - 资源类型配置

## 约束条件

### 性能指标
- gRPC 连接测试 **< 5秒** 超时
- RPC 调用执行 **< 30秒**（可配置）
- Proto 解析 **< 10秒**（大型文件）
- 支持 **100+** 并发 RPC 调用

### 安全要求
- **必须**支持 TLS 1.2+ 连接
- **必须**验证服务器证书（可配置跳过）
- **必须**限制消息大小（最大 10MB）
- **必须**在日志中脱敏敏感配置

### 可靠性要求
- **必须**实现连接池（复用连接）
- **必须**处理连接断开（自动重连）
- **必须**处理 gRPC 错误码（转换为标准化错误）
- **必须**验证服务和方法存在性（通过反射）

### 测试要求
- **必须**测试 TLS 连接（启用/跳过验证）
- **必须**测试反射 API 调用
- **必须**测试 Proto 解析（各种语法）
- **必须**测试 RPC 错误处理
- **必须**测试连接池功能

## 验收标准
- [ ] 插件符合插件开发规范
- [ ] 支持 gRPC 连接和 TLS 配置
- [ ] 支持通过反射动态调用 RPC 方法
- [ ] 支持从 Proto 文件批量导入工具
- [ ] 单元测试覆盖率 > 70%
- [ ] 可成功编译：`go build -tags grpc`

## 技术难点
- gRPC 动态调用需要反射支持
- Proto 解析和消息构建复杂度高
- 可能需要简化实现（MVP 阶段只支持静态生成的客户端）

## 参考文档
- `docs/plugin-development-guide.md` - 资源插件规范
- `docs/PRD.md` - 资源类型配置

## 预估工时
5-6 天（含反射和Proto解析）

---

## 🎯 AI生成指令

### 你必须遵守的铁律：
1. **不要**生成任何超出本任务范围的代码
2. **不要**修改或重构已有的文件（除非本任务明确要求）
3. **必须**在代码中添加关键注释说明设计决策
4. **必须**在 `init()` 中注册插件到全局注册表
5. **必须**添加接口编译时检查
6. **必须**实现所有接口方法（元数据、Schema、连接测试、工具执行）
7. **必须**返回标准化的 `PluginError` 类型
8. **必须**验证所有配置参数（serverAddress、TLS 等）
9. **必须**使用 `grpc.DialContext` 传递上下文
10. **必须**处理 gRPC 反射和动态调用

### ⚠️ 常见陷阱提醒：
- [ ] 不要忘记在 `replace_in_file` 时使用**精确的旧代码字符串**
- [ ] 不要忘记在文件顶部添加 `//go:build grpc`
- [ ] 不要忘记在 `plugins/plugins.yaml` 中注册 `grpc` 标签
- [ ] 不要忘记使用 `defer conn.Close()` 关闭连接
- [ ] 不要忘记设置连接超时（默认 30 秒）
- [ ] 不要忽略 gRPC 错误码（需要转换为标准化错误）
- [ ] 不要在日志中输出敏感配置
- [ ] 不要忘记处理 Proto 的 import 语句
- [ ] 不要忘记缓存反射结果（提高性能）
- [ ] 不要忘记验证服务和方法存在性

### gRPC 连接示例：
```go
// ✅ 正确
func (p *gRPCPlugin) createConnection(ctx context.Context, config map[string]interface{}) (*grpc.ClientConn, error) {
    serverAddress := config["serverAddress"].(string)
    
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    opts := []grpc.DialOption{
        grpc.WithBlock(),
    }
    
    // TLS 配置
    if tls, ok := config["tls"].(map[string]interface{}); ok {
        if enabled, _ := tls["enabled"].(bool); enabled {
            insecureVal, _ := tls["insecure"].(bool)
            if insecureVal {
                opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
            } else {
                // 生产环境：加载实际证书
                creds, err := credentials.NewClientTLSFromFile("cert.pem", "")
                if err != nil {
                    return nil, fmt.Errorf("load tls cert: %w", err)
                }
                opts = append(opts, grpc.WithTransportCredentials(creds))
            }
        }
    } else {
        opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
    }
    
    conn, err := grpc.DialContext(ctx, serverAddress, opts...)
    if err != nil {
        return nil, &plugin.PluginError{
            Type:    "CONNECTION_FAILED",
            Message: fmt.Sprintf("failed to connect to %s: %v", serverAddress, err),
            Cause:   plugin.ErrConnectionFailed,
        }
    }
    
    return conn, nil
}
```

### 📝 完成标准：
1. 运行测试：`GOTOOLCHAIN=local go test ./plugins/resource/grpc/... -v -race -cover`
2. 检查lint：`golangci-lint run ./plugins/resource/grpc/...`
3. 验证构建：`go build -tags grpc ./plugins/resource/grpc/`
4. 手动测试连接（如有测试 gRPC 服务器）
5. 提供简要的完成总结
6. **调用 `open_result_view` 展示主要交付文件**

### 📚 必须参考：
- `plans/00-AI生成提示词优化指南.md` - 所有强制规范
- `AGENTS.md` - 项目整体规范
- `docs/plugin-development-guide.md` - 插件开发详细规范
- 本文档开头的"项目规范提醒"章节

**关键提醒**：gRPC 插件复杂度较高，特别是反射和 Proto 解析部分。如果遇到困难，建议先实现基础功能（连接测试），再逐步添加高级功能（反射、动态调用）！
