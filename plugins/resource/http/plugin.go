package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"mengri-flow/internal/infra/plugin"
)

func init() {
	registry := plugin.GlobalRegistry()
	registry.RegisterResource(&HTTPPlugin{})
}

// HTTPPlugin HTTP资源插件
type HTTPPlugin struct{}

var _ plugin.ResourcePlugin = (*HTTPPlugin)(nil)

// PluginMeta 返回插件元数据
func (p *HTTPPlugin) PluginMeta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "http",
		Type:        plugin.PluginTypeResource,
		Version:     "1.0.0",
		Description: "HTTP/HTTPS RESTful API 资源插件",
		Author:      "mengri-flow",
		BuildTag:    "resource_http",
	}
}

// ConfigSchema 返回配置Schema
func (p *HTTPPlugin) ConfigSchema() plugin.JSONSchema {
	return plugin.NewSchemaBuilder().
		AddStringField("baseURL", "基础URL", "API的基础URL，如 https://api.example.com", true, "uri").
		AddStringField("timeout", "超时时间", "请求超时时间，如 30s", false).
		AddBooleanField("insecure", "跳过TLS验证", "是否跳过TLS证书验证（仅用于开发环境）", false, false).
		AddObjectField(
			"auth",
			"认证配置",
			"HTTP认证配置",
			map[string]any{
				"type": plugin.BuildEnumSchema(
					"认证类型",
					"认证类型",
					[]any{"none", "basic", "bearer", "apiKey"},
					"none",
				),
				"username": plugin.BuildStringSchema("用户名", "Basic认证用户名", false),
				"password": plugin.BuildStringSchema("密码", "Basic认证密码", false, "password"),
				"token":    plugin.BuildStringSchema("Token", "Bearer token或API Key", false, "password"),
				"header":   plugin.BuildStringSchema("Header名称", "API Key的Header名称", false),
			},
			[]string{"type"},
		).
		AddObjectField(
			"defaultHeaders",
			"默认Headers",
			"默认的HTTP Headers",
			map[string]any{
				"Content-Type": plugin.BuildStringSchema("Content-Type", "默认Content-Type", false),
				"User-Agent":   plugin.BuildStringSchema("User-Agent", "默认User-Agent", false),
			},
			nil,
		).
		Build()
}

// TestConnection 测试连接
func (p *HTTPPlugin) TestConnection(ctx context.Context, config map[string]any) error {
	client, err := p.createHTTPClient(config)
	if err != nil {
		return plugin.NewPluginError("connection_failed", "创建HTTP客户端失败", err)
	}

	baseURL, _ := config["baseURL"].(string)
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return plugin.NewPluginError("invalid_request", "创建请求失败", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return plugin.NewPluginError("connection_failed", "连接失败", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return plugin.NewPluginError("connection_failed", fmt.Sprintf("HTTP错误: %d", resp.StatusCode), nil)
	}

	return nil
}

// ExecuteTool 执行HTTP工具
func (p *HTTPPlugin) ExecuteTool(
	ctx context.Context,
	resourceConfig map[string]any,
	toolConfig map[string]any,
	input any,
) (*plugin.ToolResult, error) {
	startTime := time.Now()

	client, err := p.createHTTPClient(resourceConfig)
	if err != nil {
		return nil, plugin.NewPluginError("invalid_config", "配置无效", err)
	}

	// 解析工具配置
	method, _ := toolConfig["method"].(string)
	path, _ := toolConfig["path"].(string)
	baseURL, _ := resourceConfig["baseURL"].(string)

	url := baseURL + path

	// 处理请求体
	var body io.Reader
	if input != nil && method != "GET" && method != "HEAD" {
		jsonData, err := json.Marshal(input)
		if err != nil {
			return nil, plugin.NewPluginError("invalid_input", "输入数据格式错误", err)
		}
		body = bytes.NewReader(jsonData)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, plugin.NewPluginError("invalid_request", "创建请求失败", err)
	}

	// 设置认证
	if err := p.setAuth(req, resourceConfig); err != nil {
		return nil, err
	}

	// 设置默认Headers
	if headers, ok := resourceConfig["defaultHeaders"].(map[string]any); ok {
		for key, value := range headers {
			if v, ok := value.(string); ok {
				req.Header.Set(key, v)
			}
		}
	}

	// 执行请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, plugin.NewPluginError("execution_failed", "请求执行失败", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, plugin.NewPluginError("execution_failed", "读取响应失败", err)
	}

	// 解析响应数据
	var data any
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &data); err != nil {
			// 如果不是JSON，使用原始字符串
			data = string(respBody)
		}
	}

	duration := time.Since(startTime).Milliseconds()

	return &plugin.ToolResult{
		StatusCode: resp.StatusCode,
		Data:       data,
		Headers:    mapHeaders(resp.Header),
		Duration:   duration,
		Error:      "",
	}, nil
}

// ExtractTools 提取工具定义
func (p *HTTPPlugin) ExtractTools(ctx context.Context, config map[string]any) ([]plugin.ToolDefinition, error) {
	// HTTP插件不自动提取工具，需要通过UI手动配置
	return []plugin.ToolDefinition{}, nil
}

// createHTTPClient 创建HTTP客户端
func (p *HTTPPlugin) createHTTPClient(config map[string]any) (*http.Client, error) {
	timeout := 30 * time.Second
	if t, ok := config["timeout"].(string); ok && t != "" {
		if parsed, err := time.ParseDuration(t); err == nil {
			timeout = parsed
		}
	}

	insecure := false
	if i, ok := config["insecure"].(bool); ok {
		insecure = i
	}

	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: getTLSConfig(insecure),
		},
	}, nil
}

// getTLSConfig 获取TLS配置
func getTLSConfig(insecure bool) *tls.Config {
	if insecure {
		return &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	return nil
}

// setAuth 设置认证信息
func (p *HTTPPlugin) setAuth(req *http.Request, config map[string]any) error {
	auth, ok := config["auth"].(map[string]any)
	if !ok {
		return nil
	}

	authType, _ := auth["type"].(string)

	switch authType {
	case "basic":
		username, _ := auth["username"].(string)
		password, _ := auth["password"].(string)
		req.SetBasicAuth(username, password)

	case "bearer":
		token, _ := auth["token"].(string)
		req.Header.Set("Authorization", "Bearer "+token)

	case "apiKey":
		token, _ := auth["token"].(string)
		header, _ := auth["header"].(string)
		if header == "" {
			header = "X-API-Key"
		}
		req.Header.Set(header, token)
	}

	return nil
}

// mapHeaders 转换HTTP headers
func mapHeaders(header http.Header) map[string]string {
	headers := make(map[string]string)
	for key, values := range header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}
