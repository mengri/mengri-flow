# API 编排平台 — 插件开发规范

> **版本**：v1.0  
> **日期**：2026-04-01  
> **状态**：Draft  
> **作者**：软件架构师  

---

## 目录

1. [概述](#1-概述)
2. [插件类型](#2-插件类型)
3. [开发环境准备](#3-开发环境准备)
4. [资源插件开发](#4-资源插件开发)
5. [触发器插件开发](#5-触发器插件开发)
6. [配置 Schema 规范](#6-配置-schema-规范)
7. [错误处理规范](#7-错误处理规范)
8. [测试规范](#8-测试规范)
9. [最佳实践](#9-最佳实践)
10. [示例插件](#10-示例插件)

---

## 1. 概述

### 1.1 什么是插件

插件是 API 编排平台的扩展机制，允许开发者添加新的资源类型和触发器类型，而无需修改核心代码。

- **资源插件**：连接外部服务（HTTP API、gRPC 服务、数据库等），提供工具执行能力
- **触发器插件**：监听外部事件（HTTP 请求、定时任务、消息队列等），触发流程执行

### 1.2 编译时插件架构

本平台采用**编译时集成（Build-time Plugin）**模式：

- 插件源码位于 `plugins/` 目录
- 通过 Go build tags 选择性编译
- 不同客户版本可包含不同插件组合
- 单一二进制部署，无运行时依赖

### 1.3 为什么不使用 Go Plugin

| 方案 | 优点 | 缺点 |
|------|------|------|
| **Go Plugin** (`plugin` 包) | 运行时动态加载 | 版本严格匹配、跨平台差、Windows 不支持、调试困难 |
| **编译时插件** (本方案) | 类型安全、零运行时开销、单一二进制、交叉编译简单 | 新增插件需重新编译 |

---

## 2. 插件类型

### 2.1 资源插件 (ResourcePlugin)

资源插件负责：
- 管理外部服务的连接配置
- 提供连接测试能力
- 执行工具调用
- （可选）批量导入工具定义

**典型资源插件**：
- `http` - HTTP/HTTPS 服务
- `grpc` - gRPC 服务
- `postgres` - PostgreSQL 数据库
- `mysql` - MySQL 数据库
- `redis` - Redis 缓存
- `mongodb` - MongoDB 数据库

### 2.2 触发器插件 (TriggerPlugin)

触发器插件负责：
- 监听外部事件源
- 接收事件并转换为流程输入
- 管理事件订阅生命周期

**典型触发器插件**：
- `restful` - HTTP Webhook 触发
- `timer` - 定时任务触发（Cron）
- `rabbitmq` - RabbitMQ 消息触发
- `kafka` - Kafka 消息触发
- `nats` - NATS 消息触发

---

## 3. 开发环境准备

### 3.1 目录结构

```
backend/
├── cmd/server/           # 主程序入口
├── internal/             # 内部代码
│   ├── infra/plugin/     # 插件框架
│   │   ├── types.go      # 插件接口定义
│   │   ├── registry.go   # 插件注册表
│   │   └── global.go     # 全局注册表实例
│   └── ...
└── plugins/              # 插件目录 ⭐
    ├── resource/         # 资源插件
    │   ├── http/         # HTTP 资源插件
    │   ├── grpc/         # gRPC 资源插件
    │   └── postgres/     # PostgreSQL 资源插件
    └── trigger/          # 触发器插件
        ├── restful/      # RESTful 触发器
        ├── timer/        # 定时器触发器
        └── rabbitmq/     # RabbitMQ 触发器
```

### 3.2 创建新插件

```bash
# 1. 创建插件目录
mkdir -p plugins/resource/myplugin

# 2. 创建插件文件
touch plugins/resource/myplugin/plugin.go
touch plugins/resource/myplugin/plugin_test.go

# 3. 添加 build tag 到文件顶部
# //go:build myplugin
```

### 3.3 添加到构建配置

编辑 `plugins/plugins.yaml`：

```yaml
version: "1.0"
build_tags:
  - http
  - grpc
  - myplugin    # ← 添加新插件
```

---

## 4. 资源插件开发

### 4.1 接口定义

```go
// ResourcePlugin 资源插件接口
type ResourcePlugin interface {
    // PluginMeta 返回插件元数据
    PluginMeta() PluginMeta
    
    // ConfigSchema 返回资源配置表单 Schema（前端动态渲染）
    // 使用 JSON Schema 格式定义配置表单
    ConfigSchema() JSONSchema
    
    // TestConnection 测试资源连接是否可用
    // 在「测试连接」按钮点击时调用
    TestConnection(ctx context.Context, config map[string]interface{}) error
    
    // ExecuteTool 执行工具调用
    // 流程执行时，当节点使用此资源的工具时调用
    ExecuteTool(
        ctx context.Context,
        resourceConfig map[string]interface{},  // 资源配置（连接信息）
        toolConfig map[string]interface{},      // 工具配置（具体调用参数）
        input interface{},                      // 流程传入的输入数据
    ) (*ToolResult, error)
    
    // ExtractTools 从资源批量提取工具定义（可选）
    // 用于「批量导入」功能，如从 OpenAPI/Proto 文件导入
    ExtractTools(ctx context.Context, config map[string]interface{}) ([]ToolDefinition, error)
}
```

### 4.2 最小实现示例

```go
// plugins/resource/example/plugin.go
//go:build example

package example

import (
    "context"
    "fmt"
    "time"
    
    "backend/internal/infra/plugin"
)

func init() {
    registry := plugin.GlobalRegistry()
    registry.RegisterResource(&ExamplePlugin{})
}

// ExamplePlugin 示例资源插件
type ExamplePlugin struct{}

// PluginMeta 返回插件元数据
func (p *ExamplePlugin) PluginMeta() plugin.PluginMeta {
    return plugin.PluginMeta{
        Name:        "example",
        Type:        plugin.PluginTypeResource,
        Version:     "1.0.0",
        Description: "示例资源插件（用于开发参考）",
        Author:      "Platform Team",
        BuildTag:    "example",
    }
}

// ConfigSchema 资源配置表单 Schema
func (p *ExamplePlugin) ConfigSchema() plugin.JSONSchema {
    return plugin.JSONSchema{
        "type": "object",
        "required": []string{"endpoint", "apiKey"},
        "properties": map[string]interface{}{
            "endpoint": map[string]interface{}{
                "type":        "string",
                "title":       "服务端点",
                "description": "Example 服务的 API 地址",
                "format":      "uri",
                "placeholder": "https://api.example.com",
            },
            "apiKey": map[string]interface{}{
                "type":        "string",
                "title":       "API Key",
                "description": "用于认证的 API Key",
                "format":      "password",  // 前端显示为密码输入框
            },
            "timeout": map[string]interface{}{
                "type":        "integer",
                "title":       "超时时间",
                "description": "请求超时时间（秒）",
                "default":     30,
                "minimum":     1,
                "maximum":     300,
            },
        },
    }
}

// TestConnection 测试连接
func (p *ExamplePlugin) TestConnection(ctx context.Context, config map[string]interface{}) error {
    endpoint, ok := config["endpoint"].(string)
    if !ok || endpoint == "" {
        return fmt.Errorf("endpoint is required")
    }
    
    // 实现连接测试逻辑
    // 例如：发送健康检查请求
    
    return nil
}

// ExecuteTool 执行工具
func (p *ExamplePlugin) ExecuteTool(
    ctx context.Context,
    resourceConfig map[string]interface{},
    toolConfig map[string]interface{},
    input interface{},
) (*plugin.ToolResult, error) {
    start := time.Now()
    
    // 1. 解析资源配置
    endpoint := resourceConfig["endpoint"].(string)
    apiKey := resourceConfig["apiKey"].(string)
    
    // 2. 解析工具配置
    operation := toolConfig["operation"].(string)
    
    // 3. 执行实际调用
    // ... 实现具体的调用逻辑
    
    result := &plugin.ToolResult{
        StatusCode: 200,
        Data: map[string]interface{}{
            "message": fmt.Sprintf("Executed %s on %s", operation, endpoint),
            "input":   input,
        },
        Duration: time.Since(start),
    }
    
    return result, nil
}

// ExtractTools 批量提取工具（可选）
func (p *ExamplePlugin) ExtractTools(
    ctx context.Context,
    config map[string]interface{},
) ([]plugin.ToolDefinition, error) {
    // 实现从资源批量导入工具的逻辑
    // 例如：从 OpenAPI 文档解析所有 API 端点
    
    return []plugin.ToolDefinition{
        {
            Name:        "getUser",
            Type:        "example_method",
            Method:      "GET",
            Path:        "/users/{id}",
            Description: "获取用户信息",
            InputSchema: plugin.JSONSchema{
                "type": "object",
                "properties": map[string]interface{}{
                    "id": map[string]interface{}{
                        "type":        "string",
                        "description": "用户ID",
                    },
                },
            },
            OutputSchema: plugin.JSONSchema{
                "type": "object",
                "properties": map[string]interface{}{
                    "id":   map[string]interface{}{"type": "string"},
                    "name": map[string]interface{}{"type": "string"},
                },
            },
            Config: map[string]interface{}{
                "operation": "getUser",
            },
        },
    }, nil
}
```

### 4.3 工具执行上下文

工具执行时，插件可以访问以下上下文信息：

```go
// ExecuteTool 参数说明
func (p *Plugin) ExecuteTool(
    ctx context.Context,           // 包含 deadline、cancel 信号
    resourceConfig map[string]interface{},  // 资源创建时的配置
    toolConfig map[string]interface{},      // 工具定义时的配置
    input interface{},                      // 流程节点传入的数据
) (*plugin.ToolResult, error)

// ToolResult 返回结构
type ToolResult struct {
    StatusCode int               // HTTP 风格状态码 (200, 404, 500等)
    Data       interface{}       // 返回数据（会被流程引擎传递给下游节点）
    Headers    map[string]string // 可选的响应头信息
    Duration   time.Duration     // 执行耗时（自动记录到运行日志）
    Error      string            // 错误信息（失败时填写）
}
```

---

## 5. 触发器插件开发

### 5.1 接口定义

```go
// TriggerPlugin 触发器插件接口
type TriggerPlugin interface {
    // PluginMeta 返回插件元数据
    PluginMeta() PluginMeta
    
    // ConfigSchema 返回触发器配置表单 Schema
    // 定义触发器特有的配置项（如 Cron 表达式、队列名称等）
    ConfigSchema() JSONSchema
    
    // InputSchema 返回触发器输出数据的 Schema
    // 定义触发事件的数据结构，用于流程输入参数映射
    InputSchema() JSONSchema
    
    // OutputSchema 返回触发器期望的响应 Schema（可选）
    // 用于同步触发器（如 RESTful）定义响应格式
    OutputSchema() JSONSchema
    
    // Start 启动触发器监听
    // 当触发器被激活时调用，插件开始监听事件源
    Start(ctx context.Context, config map[string]interface{}, handler TriggerHandler) error
    
    // Stop 停止触发器监听
    // 当触发器被停用或系统关闭时调用
    Stop() error
}

// TriggerHandler 触发器回调函数
// 当插件收到事件时，调用此函数触发流程执行
type TriggerHandler func(ctx context.Context, input map[string]interface{}) (*TriggerResult, error)

// TriggerResult 触发处理结果
type TriggerResult struct {
    Success bool                   // 是否成功触发
    Data    map[string]interface{} // 响应数据（同步触发器用）
    Error   string                 // 错误信息
}
```

### 5.2 最小实现示例

```go
// plugins/trigger/example/plugin.go
//go:build example_trigger

package example

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "backend/internal/infra/plugin"
)

func init() {
    registry := plugin.GlobalRegistry()
    registry.RegisterTrigger(&ExampleTriggerPlugin{})
}

// ExampleTriggerPlugin 示例触发器插件
type ExampleTriggerPlugin struct {
    mu       sync.RWMutex
    handlers map[string]plugin.TriggerHandler  // triggerID -> handler
    configs  map[string]map[string]interface{} // triggerID -> config
    cancel   context.CancelFunc
}

// PluginMeta 返回插件元数据
func (p *ExampleTriggerPlugin) PluginMeta() plugin.PluginMeta {
    return plugin.PluginMeta{
        Name:        "example_trigger",
        Type:        plugin.PluginTypeTrigger,
        Version:     "1.0.0",
        Description: "示例触发器插件（用于开发参考）",
        Author:      "Platform Team",
        BuildTag:    "example_trigger",
    }
}

// ConfigSchema 触发器配置表单
func (p *ExampleTriggerPlugin) ConfigSchema() plugin.JSONSchema {
    return plugin.JSONSchema{
        "type": "object",
        "required": []string{"interval"},
        "properties": map[string]interface{}{
            "interval": map[string]interface{}{
                "type":        "integer",
                "title":       "触发间隔（秒）",
                "description": "每隔多少秒触发一次",
                "minimum":     1,
                "maximum":     3600,
                "default":     60,
            },
            "message": map[string]interface{}{
                "type":        "string",
                "title":       "触发消息",
                "description": "触发时携带的消息",
                "default":     "Hello from Example Trigger",
            },
        },
    }
}

// InputSchema 触发器输出数据 Schema
func (p *ExampleTriggerPlugin) InputSchema() plugin.JSONSchema {
    return plugin.JSONSchema{
        "type": "object",
        "properties": map[string]interface{}{
            "timestamp": map[string]interface{}{
                "type":        "string",
                "format":      "date-time",
                "description": "触发时间",
            },
            "message": map[string]interface{}{
                "type":        "string",
                "description": "触发消息",
            },
            "triggerId": map[string]interface{}{
                "type":        "string",
                "description": "触发器ID",
            },
        },
    }
}

// OutputSchema 响应 Schema（同步触发器用，本示例为异步，返回空）
func (p *ExampleTriggerPlugin) OutputSchema() plugin.JSONSchema {
    return nil
}

// Start 启动触发器监听
func (p *ExampleTriggerPlugin) Start(
    ctx context.Context,
    config map[string]interface{},
    handler plugin.TriggerHandler,
) error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    // 获取触发器ID（从 context 或 config 中获取）
    triggerID := config["triggerId"].(string)
    
    // 保存 handler 和 config
    if p.handlers == nil {
        p.handlers = make(map[string]plugin.TriggerHandler)
        p.configs = make(map[string]map[string]interface{})
    }
    p.handlers[triggerID] = handler
    p.configs[triggerID] = config
    
    // 启动后台 goroutine 监听事件
    ctx, cancel := context.WithCancel(ctx)
    p.cancel = cancel
    
    go p.run(triggerID, ctx, config, handler)
    
    return nil
}

// run 后台监听循环
func (p *ExampleTriggerPlugin) run(
    triggerID string,
    ctx context.Context,
    config map[string]interface{},
    handler plugin.TriggerHandler,
) {
    interval := time.Duration(config["interval"].(int)) * time.Second
    message := config["message"].(string)
    
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case t := <-ticker.C:
            // 构造触发事件数据
            input := map[string]interface{}{
                "timestamp": t.Format(time.RFC3339),
                "message":   message,
                "triggerId": triggerID,
            }
            
            // 调用 handler 触发流程
            result, err := handler(ctx, input)
            if err != nil {
                // 记录错误日志
                fmt.Printf("Trigger %s handler error: %v\n", triggerID, err)
            } else if !result.Success {
                fmt.Printf("Trigger %s handler failed: %s\n", triggerID, result.Error)
            }
        }
    }
}

// Stop 停止触发器监听
func (p *ExampleTriggerPlugin) Stop() error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    if p.cancel != nil {
        p.cancel()
    }
    
    p.handlers = nil
    p.configs = nil
    
    return nil
}
```

### 5.3 触发器生命周期

```
触发器创建（Console）
    │
    ▼
保存到数据库（status=inactive）
    │
    ▼
用户「激活」触发器
    │
    ▼
Executor 从 etcd 收到配置更新
    │
    ▼
调用 TriggerPlugin.Start(ctx, config, handler)
    │
    ▼
插件开始监听事件源（启动 HTTP 服务、订阅 MQ、启动定时器等）
    │
    ▼
收到事件 ──────────────────────────────────────► 调用 handler(ctx, input)
    │                                                  │
    │                                                  ▼
    │                                          流程引擎创建 Run 记录
    │                                                  │
    │◄──────────────────────────────────────── 返回 TriggerResult
    │
用户「停用」触发器 或 Executor 关闭
    │
    ▼
调用 TriggerPlugin.Stop()
    │
    ▼
插件清理资源（关闭连接、停止监听等）
```

### 5.4 同步 vs 异步触发器

| 特性 | 同步触发器 | 异步触发器 |
|------|-----------|-----------|
| **示例** | RESTful Webhook | Timer、RabbitMQ、Kafka |
| **调用方式** | 外部系统直接 HTTP 调用 | 插件主动监听事件源 |
| **响应要求** | 需要立即返回响应 | 无需响应 |
| **OutputSchema** | 必须定义，用于构造响应 | 可选 |
| **超时控制** | 受 HTTP 客户端超时限制 | 由插件内部控制 |
| **重试机制** | 由调用方决定 | 由插件或平台实现 |

**同步触发器示例（RESTful）**：

```go
// RESTful 触发器需要启动 HTTP 服务接收请求
func (p *RESTfulTriggerPlugin) Start(ctx context.Context, config map[string]interface{}, handler plugin.TriggerHandler) error {
    mux := http.NewServeMux()
    
    // 注册 webhook 端点
    path := config["path"].(string)  // 如 "/webhooks/order-created"
    mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
        // 解析请求体
        var input map[string]interface{}
        json.NewDecoder(r.Body).Decode(&input)
        
        // 调用 handler 执行流程
        result, err := handler(r.Context(), input)
        
        // 根据结果返回 HTTP 响应
        if err != nil || !result.Success {
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "success": false,
                "error": err.Error(),
            })
            return
        }
        
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(result.Data)
    })
    
    // 启动 HTTP 服务器
    server := &http.Server{Addr: ":8080", Handler: mux}
    go server.ListenAndServe()
    
    return nil
}
```

---

## 6. 配置 Schema 规范

### 6.1 JSON Schema 标准

插件使用 JSON Schema 定义配置表单，前端根据 Schema 动态渲染表单组件。

```go
// JSONSchema 类型定义
type JSONSchema = map[string]interface{}
```

### 6.2 支持的 Schema 属性

| 属性 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `type` | string | 数据类型：string/integer/number/boolean/array/object | `"type": "string"` |
| `title` | string | 表单标签 | `"title": "API Key"` |
| `description` | string | 帮助文本 | `"description": "用于认证"` |
| `default` | any | 默认值 | `"default": 30` |
| `enum` | array | 枚举选项 | `"enum": ["GET", "POST"]` |
| `format` | string | 格式提示：uri/email/password/date-time | `"format": "uri"` |
| `minimum` | number | 最小值（数字） | `"minimum": 1` |
| `maximum` | number | 最大值（数字） | `"maximum": 100` |
| `minLength` | integer | 最小长度（字符串） | `"minLength": 8` |
| `maxLength` | integer | 最大长度（字符串） | `"maxLength": 256` |
| `pattern` | string | 正则表达式 | `"pattern": "^https://"` |
| `required` | array | 必填字段列表 | `"required": ["host", "port"]` |
| `properties` | object | 对象属性定义 | 见示例 |
| `items` | object | 数组项定义 | `"items": {"type": "string"}` |

### 6.3 表单组件映射

| Schema 定义 | 前端组件 |
|------------|---------|
| `type: string, format: password` | 密码输入框（带显示/隐藏切换） |
| `type: string, format: uri` | URL 输入框（带格式校验） |
| `type: string, format: email` | 邮箱输入框 |
| `type: string, format: date-time` | 日期时间选择器 |
| `type: string, enum: [...]` | 下拉选择框 |
| `type: boolean` | 开关/复选框 |
| `type: integer/number` | 数字输入框 |
| `type: array` | 动态列表（可添加/删除） |
| `type: object` | 嵌套表单 |

### 6.4 复杂配置示例

```go
func (p *HTTPPlugin) ConfigSchema() plugin.JSONSchema {
    return plugin.JSONSchema{
        "type": "object",
        "required": []string{"baseURL"},
        "properties": map[string]interface{}{
            "baseURL": map[string]interface{}{
                "type":        "string",
                "title":       "基础 URL",
                "description": "API 的基础地址",
                "format":      "uri",
                "placeholder": "https://api.example.com",
                "pattern":     "^https?://",
            },
            "auth": map[string]interface{}{
                "type":        "object",
                "title":       "认证配置",
                "properties": map[string]interface{}{
                    "type": map[string]interface{}{
                        "type":    "string",
                        "title":   "认证类型",
                        "enum":    []string{"none", "basic", "bearer", "apiKey"},
                        "default": "none",
                    },
                    "username": map[string]interface{}{
                        "type":      "string",
                        "title":     "用户名",
                        "condition": map[string]interface{}{"auth.type": "basic"}, // 条件显示
                    },
                    "password": map[string]interface{}{
                        "type":      "string",
                        "title":     "密码",
                        "format":    "password",
                        "condition": map[string]interface{}{"auth.type": "basic"},
                    },
                    "token": map[string]interface{}{
                        "type":      "string",
                        "title":     "Token",
                        "format":    "password",
                        "condition": map[string]interface{}{"auth.type": "bearer"},
                    },
                },
            },
            "headers": map[string]interface{}{
                "type":        "array",
                "title":       "默认请求头",
                "description": "所有请求都会携带的默认头",
                "items": map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "key": map[string]interface{}{
                            "type": "string",
                            "title": "Header 名",
                        },
                        "value": map[string]interface{}{
                            "type": "string",
                            "title": "Header 值",
                        },
                    },
                },
            },
            "timeout": map[string]interface{}{
                "type":    "integer",
                "title":   "超时时间（秒）",
                "default": 30,
                "minimum": 1,
                "maximum": 300,
            },
            "retry": map[string]interface{}{
                "type":        "object",
                "title":       "重试配置",
                "properties": map[string]interface{}{
                    "maxAttempts": map[string]interface{}{
                        "type":    "integer",
                        "title":   "最大重试次数",
                        "default": 3,
                        "minimum": 0,
                        "maximum": 10,
                    },
                    "backoff": map[string]interface{}{
                        "type":    "integer",
                        "title":   "退避时间（毫秒）",
                        "default": 1000,
                    },
                },
            },
        },
    }
}
```

---

## 7. 错误处理规范

### 7.1 错误类型

插件应该使用带类型的错误，便于调用方区分处理：

```go
package plugin

import "errors"

// 预定义错误类型
var (
    // ErrConnectionFailed 连接失败（配置错误或网络问题）
    ErrConnectionFailed = errors.New("connection failed")
    
    // ErrAuthenticationFailed 认证失败（凭据错误）
    ErrAuthenticationFailed = errors.New("authentication failed")
    
    // ErrTimeout 执行超时
    ErrTimeout = errors.New("execution timeout")
    
    // ErrInvalidInput 输入参数无效
    ErrInvalidInput = errors.New("invalid input")
    
    // ErrNotFound 资源不存在
    ErrNotFound = errors.New("resource not found")
    
    // ErrRateLimited 被限流
    ErrRateLimited = errors.New("rate limited")
    
    // ErrInternal 插件内部错误
    ErrInternal = errors.New("internal error")
)

// PluginError 带类型的插件错误
type PluginError struct {
    Type    string // 错误类型标识
    Message string // 错误描述
    Cause   error  // 原始错误
}

func (e *PluginError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *PluginError) Unwrap() error {
    return e.Cause
}
```

### 7.2 错误处理示例

```go
func (p *HTTPPlugin) ExecuteTool(ctx context.Context, resourceConfig, toolConfig map[string]interface{}, input interface{}) (*plugin.ToolResult, error) {
    // 1. 参数校验
    endpoint, ok := resourceConfig["endpoint"].(string)
    if !ok || endpoint == "" {
        return nil, &plugin.PluginError{
            Type:    "INVALID_CONFIG",
            Message: "endpoint is required in resource config",
        }
    }
    
    // 2. 创建请求
    req, err := http.NewRequestWithContext(ctx, method, url, body)
    if err != nil {
        return nil, &plugin.PluginError{
            Type:    "INVALID_INPUT",
            Message: "failed to create request",
            Cause:   err,
        }
    }
    
    // 3. 执行请求
    resp, err := client.Do(req)
    if err != nil {
        // 区分错误类型
        if ctx.Err() == context.DeadlineExceeded {
            return nil, &plugin.PluginError{
                Type:    "TIMEOUT",
                Message: "request timeout",
                Cause:   plugin.ErrTimeout,
            }
        }
        if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
            return nil, &plugin.PluginError{
                Type:    "TIMEOUT",
                Message: "connection timeout",
                Cause:   plugin.ErrTimeout,
            }
        }
        return nil, &plugin.PluginError{
            Type:    "CONNECTION_FAILED",
            Message: "failed to connect to endpoint",
            Cause:   plugin.ErrConnectionFailed,
        }
    }
    defer resp.Body.Close()
    
    // 4. 处理响应
    switch resp.StatusCode {
    case http.StatusOK, http.StatusCreated:
        // 成功
    case http.StatusUnauthorized, http.StatusForbidden:
        return nil, &plugin.PluginError{
            Type:    "AUTHENTICATION_FAILED",
            Message: fmt.Sprintf("authentication failed: %s", resp.Status),
            Cause:   plugin.ErrAuthenticationFailed,
        }
    case http.StatusNotFound:
        return nil, &plugin.PluginError{
            Type:    "NOT_FOUND",
            Message: "resource not found",
            Cause:   plugin.ErrNotFound,
        }
    case http.StatusTooManyRequests:
        return nil, &plugin.PluginError{
            Type:    "RATE_LIMITED",
            Message: "rate limit exceeded",
            Cause:   plugin.ErrRateLimited,
        }
    default:
        return nil, &plugin.PluginError{
            Type:    "EXECUTION_FAILED",
            Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
        }
    }
    
    // 5. 构造结果
    return &plugin.ToolResult{
        StatusCode: resp.StatusCode,
        Data:       data,
        Duration:   duration,
    }, nil
}
```

### 7.3 错误码映射

| 插件错误类型 | HTTP 状态码 | 用户提示 |
|-------------|------------|---------|
| INVALID_CONFIG | 400 | 资源配置有误，请检查配置 |
| INVALID_INPUT | 400 | 输入参数有误 |
| CONNECTION_FAILED | 502 | 无法连接到目标服务 |
| AUTHENTICATION_FAILED | 401 | 认证失败，请检查凭据 |
| TIMEOUT | 504 | 请求超时 |
| NOT_FOUND | 404 | 请求的资源不存在 |
| RATE_LIMITED | 429 | 请求过于频繁，被限流 |
| EXECUTION_FAILED | 500 | 执行失败 |
| INTERNAL_ERROR | 500 | 插件内部错误 |

---

## 8. 测试规范

### 8.1 单元测试

每个插件必须包含单元测试：

```go
// plugins/resource/http/plugin_test.go
//go:build http

package http

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestHTTPPlugin_PluginMeta(t *testing.T) {
    p := &HTTPPlugin{}
    meta := p.PluginMeta()
    
    assert.Equal(t, "http", meta.Name)
    assert.Equal(t, plugin.PluginTypeResource, meta.Type)
    assert.NotEmpty(t, meta.Version)
}

func TestHTTPPlugin_ConfigSchema(t *testing.T) {
    p := &HTTPPlugin{}
    schema := p.ConfigSchema()
    
    // 验证 Schema 结构
    assert.Equal(t, "object", schema["type"])
    properties, ok := schema["properties"].(map[string]interface{})
    require.True(t, ok)
    assert.Contains(t, properties, "baseURL")
    assert.Contains(t, properties, "timeout")
}

func TestHTTPPlugin_TestConnection(t *testing.T) {
    // 创建测试服务器
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
    }))
    defer server.Close()
    
    p := &HTTPPlugin{}
    config := map[string]interface{}{
        "baseURL": server.URL,
        "timeout": 30,
    }
    
    err := p.TestConnection(context.Background(), config)
    assert.NoError(t, err)
}

func TestHTTPPlugin_ExecuteTool(t *testing.T) {
    // 创建测试服务器
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "POST", r.Method)
        assert.Equal(t, "/api/users", r.URL.Path)
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        w.Write([]byte(`{"id": "123", "name": "John"}`))
    }))
    defer server.Close()
    
    p := &HTTPPlugin{}
    
    resourceConfig := map[string]interface{}{
        "baseURL": server.URL,
        "timeout": 30,
    }
    
    toolConfig := map[string]interface{}{
        "method": "POST",
        "path":   "/api/users",
    }
    
    input := map[string]interface{}{
        "name": "John",
    }
    
    result, err := p.ExecuteTool(context.Background(), resourceConfig, toolConfig, input)
    
    require.NoError(t, err)
    assert.Equal(t, 201, result.StatusCode)
    
    data, ok := result.Data.(map[string]interface{})
    require.True(t, ok)
    assert.Equal(t, "123", data["id"])
}

func TestHTTPPlugin_ExecuteTool_Timeout(t *testing.T) {
    // 创建慢速服务器
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(2 * time.Second)
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()
    
    p := &HTTPPlugin{}
    
    resourceConfig := map[string]interface{}{
        "baseURL": server.URL,
        "timeout": 1, // 1 秒超时
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()
    
    _, err := p.ExecuteTool(ctx, resourceConfig, map[string]interface{}{}, nil)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "timeout")
}
```

### 8.2 集成测试

```go
// plugins/resource/http/plugin_integration_test.go
//go:build http && integration

package http

import (
    "context"
    "os"
    "testing"
)

// 使用真实的外部服务进行测试
// 运行: go test -tags="http integration" ./plugins/resource/http/

func TestHTTPPlugin_RealAPI(t *testing.T) {
    if os.Getenv("INTEGRATION_TEST") != "true" {
        t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
    }
    
    p := &HTTPPlugin{}
    
    config := map[string]interface{}{
        "baseURL": os.Getenv("TEST_API_URL"),
        "auth": map[string]interface{}{
            "type":  "bearer",
            "token": os.Getenv("TEST_API_TOKEN"),
        },
    }
    
    // 测试真实连接
    err := p.TestConnection(context.Background(), config)
    require.NoError(t, err)
}
```

### 8.3 Mock 工具

```go
// internal/test/plugin_mock.go

package test

import (
    "context"
    "backend/internal/infra/plugin"
)

// MockResourcePlugin 用于测试的模拟资源插件
type MockResourcePlugin struct {
    Meta            plugin.PluginMeta
    TestConnFunc    func(ctx context.Context, config map[string]interface{}) error
    ExecuteToolFunc func(ctx context.Context, resourceConfig, toolConfig map[string]interface{}, input interface{}) (*plugin.ToolResult, error)
}

func (m *MockResourcePlugin) PluginMeta() plugin.PluginMeta {
    return m.Meta
}

func (m *MockResourcePlugin) ConfigSchema() plugin.JSONSchema {
    return plugin.JSONSchema{"type": "object"}
}

func (m *MockResourcePlugin) TestConnection(ctx context.Context, config map[string]interface{}) error {
    if m.TestConnFunc != nil {
        return m.TestConnFunc(ctx, config)
    }
    return nil
}

func (m *MockResourcePlugin) ExecuteTool(ctx context.Context, resourceConfig, toolConfig map[string]interface{}, input interface{}) (*plugin.ToolResult, error) {
    if m.ExecuteToolFunc != nil {
        return m.ExecuteToolFunc(ctx, resourceConfig, toolConfig, input)
    }
    return &plugin.ToolResult{StatusCode: 200, Data: map[string]interface{}{}}, nil
}
```

---

## 9. 最佳实践

### 9.1 资源插件最佳实践

1. **连接复用**
   - 对于需要保持连接的资源（数据库、消息队列），在插件内部维护连接池
   - 不要每次 ExecuteTool 都新建连接

2. **超时控制**
   - 始终使用 `context.WithTimeout` 或 `context.WithDeadline`
   - 尊重用户配置的超时时间

3. **资源清理**
   - 实现 `io.Closer` 接口（如果适用）
   - 在插件停用或系统关闭时清理资源

4. **配置验证**
   - 在 `TestConnection` 中验证所有必需配置
   - 提供清晰的错误信息

5. **敏感信息处理**
   - 密码、Token 等敏感字段使用 `"format": "password"`
   - 日志中不要打印敏感信息

### 9.2 触发器插件最佳实践

1. **优雅关闭**
   - `Stop()` 方法必须等待所有进行中的处理完成
   - 使用 `sync.WaitGroup` 跟踪活跃的 handler 调用

2. **错误恢复**
   - 使用 `recover()` 防止 panic 导致整个触发器崩溃
   - 记录错误并继续监听

3. **限流与防抖**
   - 对于高频事件源，实现限流或防抖机制
   - 避免短时间内触发大量流程

4. **幂等性**
   - 触发器应该能够处理重复事件（如果可能）
   - 在事件数据中包含唯一标识

5. **上下文传递**
   - 将 trace ID 等信息注入 context
   - 便于日志追踪和问题排查

### 9.3 性能优化

```go
// 1. 使用对象池减少 GC 压力
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func (p *Plugin) ExecuteTool(...) (*plugin.ToolResult, error) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer bufferPool.Put(buf)
    buf.Reset()
    // 使用 buf...
}

// 2. 预编译正则表达式
var pathParamRegex = regexp.MustCompile(`\{(\w+)\}`)

// 3. 缓存 Schema（如果计算成本高）
var schemaCache = sync.Map{}

func (p *Plugin) ConfigSchema() plugin.JSONSchema {
    if cached, ok := schemaCache.Load(p.Name()); ok {
        return cached.(plugin.JSONSchema)
    }
    schema := p.buildSchema()
    schemaCache.Store(p.Name(), schema)
    return schema
}
```

### 9.4 日志规范

```go
import "backend/internal/infra/logger"

func (p *Plugin) ExecuteTool(ctx context.Context, ...) (*plugin.ToolResult, error) {
    log := logger.FromContext(ctx).With(
        "plugin", p.Name(),
        "tool", toolConfig["name"],
    )
    
    log.Debug("executing tool", "input", input)
    
    result, err := p.doExecute(ctx, ...)
    if err != nil {
        log.Error("tool execution failed", "error", err)
        return nil, err
    }
    
    log.Info("tool executed", "duration", result.Duration, "status", result.StatusCode)
    return result, nil
}
```

---

## 10. 示例插件

### 10.1 HTTP 资源插件（完整实现）

见 `plugins/resource/http/` 目录（实际开发时提供）

### 10.2 定时器触发器插件（完整实现）

见 `plugins/trigger/timer/` 目录（实际开发时提供）

### 10.3 插件模板生成器

提供脚手架命令快速创建新插件：

```bash
# 生成资源插件模板
make plugin-template RESOURCE=myplugin

# 生成触发器插件模板
make plugin-template TRIGGER=mytrigger
```

---

## 附录 A：插件检查清单

在提交插件代码前，请确认：

- [ ] 文件顶部包含正确的 `//go:build` 标签
- [ ] 实现了对应接口的所有方法
- [ ] 提供了完整的 `ConfigSchema()`，包含 `title` 和 `description`
- [ ] 敏感字段使用 `"format": "password"`
- [ ] 实现了 `TestConnection()` 用于连接测试
- [ ] 错误处理使用了带类型的错误
- [ ] 包含单元测试，覆盖率 > 70%
- [ ] 包含集成测试（如适用）
- [ ] 更新了 `plugins/plugins.yaml`
- [ ] 文档已更新（如需要）

---

## 附录 B：常见问题

**Q: 插件如何访问数据库？**

A: 插件不应该直接访问平台数据库。如需存储数据，通过接口向平台请求。

**Q: 插件可以使用第三方库吗？**

A: 可以，但请谨慎引入依赖。优先使用标准库。

**Q: 如何调试插件？**

A: 使用 `go run -tags "http timer" ./cmd/server/` 启动带指定插件的开发服务器。

**Q: 插件可以调用其他插件吗？**

A: 不建议。插件应该独立工作。如需复用逻辑，提取到公共包。

---

*本文档与架构设计文档 v1.3 保持一致*
