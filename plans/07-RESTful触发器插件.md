# 任务 07: RESTful触发器插件

## 任务概述
实现RESTful触发器插件，支持通过HTTP Webhook触发流程执行，提供同步和异步两种接口范式。

## 上下文依赖
- 任务 03: 插件框架核心
- 任务 02: 数据库实体（Trigger, Run）
- 任务 18: Executor执行器核心（流程引擎调用）

## 涉及文件
- `plugins/trigger/restful/plugin.go` - RESTful触发器插件
- `plugins/trigger/restful/plugin_test.go` - 单元测试
- `internal/infra/trigger/manager.go` - 触发器管理器（共享）

## 详细步骤

### 7.1 插件基本结构
**文件：`plugins/trigger/restful/plugin.go`**
```go
//go:build restful

package restful

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    
    "backend/internal/infra/plugin"
)

func init() {
    registry := plugin.GlobalRegistry()
    registry.RegisterTrigger(&RESTfulTriggerPlugin{})
}

type RESTfulTriggerPlugin struct {
    mu       sync.RWMutex
    servers  map[string]*http.Server  // triggerID -> server
    configs  map[string]map[string]interface{}
    handlers map[string]plugin.TriggerHandler
}
```
- [ ] 创建插件文件，添加 `//go:build restful` 构建标签
- [ ] 实现 `init()` 注册插件
- [ ] 定义结构体，包含服务器映射

### 7.2 实现插件元数据
```go
func (p *RESTfulTriggerPlugin) PluginMeta() plugin.PluginMeta {
    return plugin.PluginMeta{
        Name:        "restful",
        Type:        plugin.PluginTypeTrigger,
        Version:     "1.0.0",
        Description: "RESTful Webhook触发器插件，支持同步和异步接口",
        Author:      "Platform Team",
        BuildTag:    "restful",
    }
}
```

### 7.3 实现配置Schema
```go
func (p *RESTfulTriggerPlugin) ConfigSchema() plugin.JSONSchema {
    return plugin.JSONSchema{
        "type": "object",
        "required": []string{"path", "method"},
        "properties": map[string]interface{}{
            "path": map[string]interface{}{
                "type":        "string",
                "title":       "请求路径",
                "description": "Webhook路径，如 /webhooks/order-created",
                "pattern":     "^/",
                "placeholder": "/webhooks/your-event",
            },
            "method": map[string]interface{}{
                "type":        "string",
                "title":       "HTTP方法",
                "enum":        []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
                "default":     "POST",
            },
            "async": map[string]interface{}{
                "type":        "boolean",
                "title":       "异步模式",
                "description": "启用后返回202 Accepted，流程在后台执行",
                "default":     false,
            },
            "auth": map[string]interface{}{
                "type": "object",
                "title": "认证配置",
                "properties": map[string]interface{}{
                    "type": map[string]interface{}{
                        "type":        "string",
                        "title":       "认证类型",
                        "enum":        []string{"none", "apiKey"},
                        "default":     "none",
                    },
                    "apiKey": map[string]interface{}{
                        "type":        "string",
                        "title":       "API Key",
                        "format":      "password",
                        "condition":   map[string]interface{}{"auth.type": "apiKey"},
                    },
                    "apiKeyLocation": map[string]interface{}{
                        "type":        "string",
                        "title":       "API Key位置",
                        "enum":        []string{"header", "query"},
                        "default":     "header",
                        "condition":   map[string]interface{}{"auth.type": "apiKey"},
                    },
                    "apiKeyName": map[string]interface{}{
                        "type":        "string",
                        "title":       "API Key名称",
                        "default":     "X-API-Key",
                        "condition":   map[string]interface{}{"auth.type": "apiKey"},
                    },
                },
            },
        },
    }
}
```
- [ ] 包含路径和HTTP方法配置
- [ ] 包含同步/异步模式选择
- [ ] 包含API Key认证配置

### 7.4 实现输入输出Schema
```go
func (p *RESTfulTriggerPlugin) InputSchema() plugin.JSONSchema {
    return plugin.JSONSchema{
        "type": "object",
        "properties": map[string]interface{}{
            "headers": map[string]interface{}{
                "type":        "object",
                "description": "HTTP请求头",
            },
            "query": map[string]interface{}{
                "type":        "object",
                "description": "URL查询参数",
            },
            "body": map[string]interface{}{
                "type":        "object",
                "description": "请求体（JSON）",
            },
            "pathParams": map[string]interface{}{
                "type":        "object",
                "description": "路径参数",
            },
        },
    }
}

func (p *RESTfulTriggerPlugin) OutputSchema() plugin.JSONSchema {
    return plugin.JSONSchema{
        "type": "object",
        "properties": map[string]interface{}{
            "success": map[string]interface{}{
                "type":    "boolean",
                "description": "是否成功",
            },
            "data": map[string]interface{}{
                "type":        "object",
                "description": "响应数据",
            },
            "error": map[string]interface{}{
                "type":        "string",
                "description": "错误信息",
            },
        },
    }
}
```

### 7.5 实现Start方法（启动HTTP服务器）
```go
func (p *RESTfulTriggerPlugin) Start(
    ctx context.Context,
    config map[string]interface{},
    handler plugin.TriggerHandler,
) error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    // 获取触发器ID
    triggerID := config["triggerId"].(string)
    
    // 保存handler和配置
    if p.servers == nil {
        p.servers = make(map[string]*http.Server)
        p.configs = make(map[string]map[string]interface{})
        p.handlers = make(map[string]plugin.TriggerHandler)
    }
    
    p.configs[triggerID] = config
    p.handlers[triggerID] = handler
    
    // 创建HTTP处理器
    mux := http.NewServeMux()
    
    path := config["path"].(string)
    method := config["method"].(string)
    
    mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
        // 验证HTTP方法
        if r.Method != method {
            w.WriteHeader(http.StatusMethodNotAllowed)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "success": false,
                "error":   "method not allowed",
            })
            return
        }
        
        // 验证认证
        if err := p.validateAuth(r, config); err != nil {
            w.WriteHeader(http.StatusUnauthorized)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "success": false,
                "error":   "unauthorized",
            })
            return
        }
        
        // 解析请求
        input, err := p.parseRequest(r)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "success": false,
                "error":   fmt.Sprintf("invalid request: %v", err),
            })
            return
        }
        
        // 检查是否为异步模式
        async := false
        if val, ok := config["async"].(bool); ok {
            async = val
        }
        
        if async {
            // 异步模式：返回202，后台执行
            w.WriteHeader(http.StatusAccepted)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "success": true,
                "message": "request accepted, processing asynchronously",
            })
            
            // 在后台goroutine中执行
            go func() {
                p.executeInBackground(ctx, triggerID, input)
            }()
        } else {
            // 同步模式：等待执行完成
            result, err := handler(ctx, input)
            if err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                json.NewEncoder(w).Encode(map[string]interface{}{
                    "success": false,
                    "error":   err.Error(),
                })
                return
            }
            
            if result.Success {
                w.WriteHeader(http.StatusOK)
                json.NewEncoder(w).Encode(map[string]interface{}{
                    "success": true,
                    "data":    result.Data,
                })
            } else {
                w.WriteHeader(http.StatusBadRequest)
                json.NewEncoder(w).Encode(map[string]interface{}{
                    "success": false,
                    "error":   result.Error,
                })
            }
        }
    })
    
    // 创建HTTP服务器
    server := &http.Server{
        Addr:    ":8080",  // 实际端口应从配置获取
        Handler: mux,
    }
    
    p.servers[triggerID] = server
    
    // 在后台启动服务器
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            fmt.Printf("RESTful trigger server error: %v\n", err)
        }
    }()
    
    return nil
}
```
- [ ] 实现Start方法，启动HTTP服务器
- [ ] 支持路径和方法验证
- [ ] 支持API Key认证验证
- [ ] 实现同步模式（等待流程执行完成）
- [ ] 实现异步模式（返回202，后台执行）

### 7.6 辅助方法实现
```go
// validateAuth 验证请求认证
func (p *RESTfulTriggerPlugin) validateAuth(r *http.Request, config map[string]interface{}) error

// parseRequest 解析HTTP请求为输入数据
func (p *RESTfulTriggerPlugin) parseRequest(r *http.Request) (map[string]interface{}, error)

// executeInBackground 在后台执行流程
func (p *RESTfulTriggerPlugin) executeInBackground(ctx context.Context, triggerID string, input map[string]interface{})
```

### 7.7 实现Stop方法
```go
func (p *RESTfulTriggerPlugin) Stop() error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    for triggerID, server := range p.servers {
        // 优雅关闭HTTP服务器
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        if err := server.Shutdown(shutdownCtx); err != nil {
            fmt.Printf("Failed to shutdown server for trigger %s: %v\n", triggerID, err)
        }
    }
    
    // 清空映射
    p.servers = nil
    p.configs = nil
    p.handlers = nil
    
    return nil
}
```
- [ ] 实现Stop方法，优雅关闭所有HTTP服务器
- [ ] 5秒超时限制
- [ ] 清理资源

### 7.8 端口和服务器管理
**问题：** 多个触发器不能都监听:8080
**解决方案：**
```go
// 方案1：使用统一网关服务器，按路径分发
type GatewayServer struct {
    mu       sync.RWMutex
    triggers map[string]*TriggerRoute  // path -> trigger
}

// 方案2：每个触发器分配不同端口（从端口池分配）
```
- [ ] 设计并实现多触发器端口管理方案
- [ ] 推荐方案1：统一网关，按路径路由到不同触发器处理器

### 7.9 单元测试
**文件：`plugins/trigger/restful/plugin_test.go`**
```go
func TestRESTfulTriggerPlugin_PluginMeta(t *testing.T)
func TestRESTfulTriggerPlugin_ConfigSchema(t *testing.T)
func TestRESTfulTriggerPlugin_Start_Stop(t *testing.T)
func TestRESTfulTriggerPlugin_SyncExecution(t *testing.T)
func TestRESTfulTriggerPlugin_AsyncExecution(t *testing.T)
func TestRESTfulTriggerPlugin_AuthValidation(t *testing.T)
```
- [ ] 使用 httptest 测试HTTP处理器
- [ ] 测试同步和异步模式
- [ ] 测试认证验证
- [ ] 测试错误处理

### 7.10 更新插件配置
**文件：`plugins/plugins.yaml`**
- [ ] 添加 `restful` 到 build_tags

## 验收标准
- [ ] 插件符合插件开发规范
- [ ] 支持同步和异步两种模式
- [ ] 支持API Key认证
- [ ] 支持自定义路径和HTTP方法
- [ ] 单元测试覆盖率 > 75%
- [ ] 可成功编译：`go build -tags restful`
- [ ] 能正确接收HTTP请求并触发流程
- [ ] 同步模式返回流程执行结果
- [ ] 异步模式返回202 Accepted

## 设计决策
- **同步vs异步：** 由触发器配置决定，非请求参数
- **端口管理：** 使用统一网关服务器，避免端口冲突
- **认证方式：** 优先支持API Key，后续可扩展OAuth/JWT

## 参考文档
- `docs/plugin-development-guide.md` - 第5章触发器插件开发
- `docs/plugin-development-guide.md` - 5.4同步vs异步触发器
- `docs/PRD.md` - 触发器类型配置

## 预估工时
4-5 天（含端口管理和测试）
