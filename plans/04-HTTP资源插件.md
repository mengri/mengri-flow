# 任务 04: HTTP资源插件

## 任务概述
实现HTTP/HTTPS资源插件，支持RESTful API的连接管理、工具执行和批量导入（OpenAPI/Swagger）。

## 上下文依赖
- 任务 03: 插件框架核心

## 涉及文件
- `plugins/resource/http/plugin.go` - HTTP插件实现
- `plugins/resource/http/plugin_test.go` - 单元测试
- `plugins/resource/http/openapi_parser.go` - OpenAPI解析器

## 详细步骤

### 4.1 插件基本结构
**文件：`plugins/resource/http/plugin.go`**
```go
//go:build http

package http

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
    
    "backend/internal/infra/plugin"
)

func init() {
    registry := plugin.GlobalRegistry()
    registry.RegisterResource(&HTTPPlugin{})
}

type HTTPPlugin struct{}
```
- [ ] 创建插件文件，添加 `//go:build http` 构建标签
- [ ] 实现 `init()` 注册插件
- [ ] 定义 `HTTPPlugin` 结构体

### 4.2 实现插件元数据
```go
func (p *HTTPPlugin) PluginMeta() plugin.PluginMeta {
    return plugin.PluginMeta{
        Name:        "http",
        Type:        plugin.PluginTypeResource,
        Version:     "1.0.0",
        Description: "HTTP/HTTPS RESTful API 资源插件",
        Author:      "Platform Team",
        BuildTag:    "http",
    }
}
```
- [ ] 返回插件元数据

### 4.3 实现配置Schema
```go
func (p *HTTPPlugin) ConfigSchema() plugin.JSONSchema {
    return plugin.JSONSchema{
        "type": "object",
        "required": []string{"baseURL"},
        "properties": map[string]interface{}{
            "baseURL": map[string]interface{}{
                "type":        "string",
                "title":       "基础 URL",
                "description": "API的基础地址，如 https://api.example.com",
                "format":      "uri",
                "pattern":     "^https?://",
            },
            "auth": map[string]interface{}{
                "type": "object",
                "title": "认证配置",
                "properties": map[string]interface{}{
                    "type": map[string]interface{}{
                        "type":        "string",
                        "title":       "认证类型",
                        "enum":        []string{"none", "basic", "bearer", "apiKey"},
                        "default":     "none",
                    },
                    "username": map[string]interface{}{
                        "type":        "string",
                        "title":       "用户名",
                        "condition":   map[string]interface{}{"auth.type": "basic"},
                    },
                    "password": map[string]interface{}{
                        "type":        "string",
                        "title":       "密码",
                        "format":      "password",
                        "condition":   map[string]interface{}{"auth.type": "basic"},
                    },
                    "token": map[string]interface{}{
                        "type":        "string",
                        "title":       "Token",
                        "format":      "password",
                        "condition":   map[string]interface{}{"auth.type": "bearer"},
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
            "headers": map[string]interface{}{
                "type":        "array",
                "title":       "默认请求头",
                "description": "所有请求都会携带的默认头",
                "items": map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "key":   map[string]interface{}{"type": "string"},
                        "value": map[string]interface{}{"type": "string"},
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
- [ ] 实现完整的配置Schema
- [ ] 包含认证配置（basic/bearer/apiKey）
- [ ] 包含超时、重试、默认头配置

### 4.4 实现连接测试
```go
func (p *HTTPPlugin) TestConnection(ctx context.Context, config map[string]interface{}) error {
    baseURL, ok := config["baseURL"].(string)
    if !ok || baseURL == "" {
        return &plugin.PluginError{
            Type:    "INVALID_CONFIG",
            Message: "baseURL is required",
        }
    }
    
    client := p.createHTTPClient(config)
    
    // 尝试访问根路径或健康检查端点
    healthURL := baseURL + "/health"
    req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
    if err != nil {
        return fmt.Errorf("create request: %w", err)
    }
    
    p.setAuthHeaders(req, config)
    
    resp, err := client.Do(req)
    if err != nil {
        return &plugin.PluginError{
            Type:    "CONNECTION_FAILED",
            Message: fmt.Sprintf("failed to connect to %s", baseURL),
            Cause:   plugin.ErrConnectionFailed,
        }
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        return &plugin.PluginError{
            Type:    "AUTHENTICATION_FAILED",
            Message: fmt.Sprintf("health check failed with status %d", resp.StatusCode),
            Cause:   plugin.ErrAuthenticationFailed,
        }
    }
    
    return nil
}
```
- [ ] 实现连接测试逻辑
- [ ] 支持自定义超时
- [ ] 返回标准化错误类型

### 4.5 实现工具执行
```go
func (p *HTTPPlugin) ExecuteTool(
    ctx context.Context,
    resourceConfig map[string]interface{},
    toolConfig map[string]interface{},
    input interface{},
) (*plugin.ToolResult, error) {
    start := time.Now()
    
    // 1. 解析资源配置
    baseURL := resourceConfig["baseURL"].(string)
    timeout := p.getTimeout(resourceConfig)
    
    // 2. 解析工具配置
    method := toolConfig["method"].(string)
    path := toolConfig["path"].(string)
    
    // 3. 处理输入参数（替换路径参数和请求体）
    resolvedPath := p.resolvePathParams(path, input)
    body, err := p.buildRequestBody(toolConfig, input)
    if err != nil {
        return nil, err
    }
    
    // 4. 创建HTTP请求
    client := p.createHTTPClientWithTimeout(timeout)
    req, err := http.NewRequestWithContext(ctx, method, baseURL+resolvedPath, body)
    if err != nil {
        return nil, err
    }
    
    // 5. 设置认证和头信息
    p.setAuthHeaders(req, resourceConfig)
    p.setDefaultHeaders(req, resourceConfig)
    
    // 6. 执行请求
    resp, err := client.Do(req)
    if err != nil {
        return p.handleError(err)
    }
    defer resp.Body.Close()
    
    // 7. 读取响应
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    // 8. 解析响应
    var data interface{}
    if err := json.Unmarshal(respBody, &data); err != nil {
        data = string(respBody) // 非JSON响应作为字符串处理
    }
    
    duration := time.Since(start)
    
    return &plugin.ToolResult{
        StatusCode: resp.StatusCode,
        Data:       data,
        Headers:    resp.Header,
        Duration:   duration,
    }, nil
}
```
- [ ] 实现工具执行逻辑
- [ ] 支持路径参数替换（如 `/users/{id}`）
- [ ] 支持请求体构建
- [ ] 支持所有认证方式
- [ ] 支持超时控制

### 4.6 辅助方法实现
```go
// createHTTPClient 创建HTTP客户端
func (p *HTTPPlugin) createHTTPClient(config map[string]interface{}) *http.Client

// setAuthHeaders 设置认证头
func (p *HTTPPlugin) setAuthHeaders(req *http.Request, config map[string]interface{})

// resolvePathParams 解析路径参数
func (p *HTTPPlugin) resolvePathParams(path string, input interface{}) string

// buildRequestBody 构建请求体
func (p *HTTPPlugin) buildRequestBody(toolConfig map[string]interface{}, input interface{}) (io.Reader, error)

// handleError 处理执行错误并返回标准化错误类型
func (p *HTTPPlugin) handleError(err error) (*plugin.ToolResult, error)
```
- [ ] 实现所有辅助方法
- [ ] 正确处理各种输入类型

### 4.7 OpenAPI批量导入
**文件：`plugins/resource/http/openapi_parser.go`**
```go
// parseOpenAPI 解析OpenAPI/Swagger文档
func parseOpenAPI(data []byte) ([]plugin.ToolDefinition, error)
```
- [ ] 实现OpenAPI v3解析器
- [ ] 从paths中提取所有operation
- [ ] 转换parameters和requestBody为InputSchema
- [ ] 转换responses为OutputSchema
- [ ] 支持$ref引用解析

```go
func (p *HTTPPlugin) ExtractTools(ctx context.Context, config map[string]interface{}) ([]plugin.ToolDefinition, error) {
    baseURL := config["baseURL"].(string)
    
    // 尝试获取OpenAPI文档
    openapiURL := baseURL + "/openapi.json"
    
    client := p.createHTTPClient(config)
    resp, err := client.Get(openapiURL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    data, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    return parseOpenAPI(data)
}
```
- [ ] 实现ExtractTools方法
- [ ] 从OpenAPI端点自动发现工具

### 4.8 单元测试
**文件：`plugins/resource/http/plugin_test.go`**
```go
func TestHTTPPlugin_PluginMeta(t *testing.T)
func TestHTTPPlugin_ConfigSchema(t *testing.T)
func TestHTTPPlugin_TestConnection_Success(t *testing.T)
func TestHTTPPlugin_TestConnection_Failure(t *testing.T)
func TestHTTPPlugin_ExecuteTool_GET(t *testing.T)
func TestHTTPPlugin_ExecuteTool_POST(t *testing.T)
func TestHTTPPlugin_ExecuteTool_WithPathParams(t *testing.T)
func TestHTTPPlugin_ExecuteTool_WithAuth(t *testing.T)
func TestHTTPPlugin_ExecuteTool_Timeout(t *testing.T)
func TestParseOpenAPI(t *testing.T)
```
- [ ] 使用 httptest 创建测试服务器
- [ ] 测试所有认证类型
- [ ] 测试超时和错误处理
- [ ] 测试OpenAPI解析
- [ ] 覆盖率 > 80%

### 4.9 更新插件配置
**文件：`plugins/plugins.yaml`**
```yaml
version: "1.0"
build_tags:
  - http  # 添加http插件
  - example
  - example_trigger
```
- [ ] 将http添加到构建配置

## 验收标准
- [ ] 插件符合插件开发规范
- [ ] 所有接口方法实现完整
- [ ] 支持基本认证、Bearer Token、API Key三种认证方式
- [ ] 支持OpenAPI v3批量导入
- [ ] 单元测试覆盖率 > 80%
- [ ] 使用 `go build -tags http` 可成功编译
- [ ] 连接测试准确反映服务状态
- [ ] 错误处理返回标准化错误类型

## 参考文档
- `docs/plugin-development-guide.md` - 第4章资源插件开发
- `docs/plugin-development-guide.md` - 第6章配置Schema规范
- `docs/plugin-development-guide.md` - 第7章错误处理规范
- `docs/plugin-development-guide.md` - 第8章测试规范

## 预估工时
4-5 天
