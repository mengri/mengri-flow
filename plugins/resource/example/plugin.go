package example

import (
	"context"
	"fmt"
	"time"

	"mengri-flow/internal/infra/plugin"
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
		"type":     "object",
		"required": []string{"endpoint"},
		"properties": map[string]any{
			"endpoint": map[string]any{
				"type":        "string",
				"title":       "服务端点",
				"description": "Example 服务的 API 地址",
				"format":      "uri",
				"placeholder": "https://api.example.com",
			},
			"apiKey": map[string]any{
				"type":        "string",
				"title":       "API Key",
				"description": "用于认证的 API Key",
				"format":      "password",
			},
		},
	}
}

// TestConnection 测试连接
func (p *ExamplePlugin) TestConnection(ctx context.Context, config map[string]any) error {
	endpoint, ok := config["endpoint"].(string)
	if !ok || endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}
	return nil
}

// ExecuteTool 执行工具
func (p *ExamplePlugin) ExecuteTool(
	ctx context.Context,
	resourceConfig map[string]any,
	toolConfig map[string]any,
	input any,
) (*plugin.ToolResult, error) {
	start := time.Now()

	endpoint := resourceConfig["endpoint"].(string)
	operation := toolConfig["operation"].(string)

	result := &plugin.ToolResult{
		StatusCode: 200,
		Data: map[string]any{
			"message": fmt.Sprintf("Executed %s on %s", operation, endpoint),
			"input":   input,
		},
		Duration: time.Since(start),
	}

	return result, nil
}

// ExtractTools 批量提取工具（可选）
func (p *ExamplePlugin) ExtractTools(ctx context.Context, config map[string]any) ([]plugin.ToolDefinition, error) {
	return []plugin.ToolDefinition{
		{
			Name:        "getUser",
			Type:        "example_method",
			Method:      "GET",
			Path:        "/users/{id}",
			Description: "获取用户信息",
		},
	}, nil
}
