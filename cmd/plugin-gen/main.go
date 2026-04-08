package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	pluginType = flag.String("type", "", "插件类型: resource 或 trigger")
	pluginName = flag.String("name", "", "插件名称")
	outputDir  = flag.String("output", "./plugins", "输出目录")
)

func main() {
	flag.Parse()

	if *pluginType == "" || *pluginName == "" {
		fmt.Println("Usage: plugin-gen --type=<resource|trigger> --name=<plugin-name>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *pluginType != "resource" && *pluginType != "trigger" {
		fmt.Printf("无效的插件类型: %s, 必须是 resource 或 trigger\n", *pluginType)
		os.Exit(1)
	}

	if err := generatePlugin(*pluginType, *pluginName, *outputDir); err != nil {
		fmt.Printf("生成插件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("成功生成 %s 插件: %s\n", *pluginType, *pluginName)
}

func generatePlugin(pluginType, name, output string) error {
	// 创建插件目录
	pluginDir := filepath.Join(output, pluginType+"_"+name)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("创建插件目录失败: %w", err)
	}

	// 生成插件文件
	data := map[string]string{
		"PluginType":    pluginType,
		"PluginName":    name,
		"PackageName":   name,
		"BuildTag":      pluginType + "_" + name,
		"InterfaceName": strings.Title(pluginType) + "Plugin",
		"TypeName":      strings.Title(name) + strings.Title(pluginType),
	}

	// 生成主插件文件
	if err := generatePluginFile(pluginDir, "plugin.go", pluginTemplate, data); err != nil {
		return err
	}

	// 生成测试文件
	if err := generatePluginFile(pluginDir, "plugin_test.go", testTemplate, data); err != nil {
		return err
	}

	return nil
}

func generatePluginFile(dir, filename, tmpl string, data map[string]string) error {
	filePath := filepath.Join(dir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件 %s 失败: %w", filename, err)
	}
	defer file.Close()

	t := template.Must(template.New("plugin").Parse(tmpl))
	if err := t.Execute(file, data); err != nil {
		return fmt.Errorf("生成文件 %s 失败: %w", filename, err)
	}

	return nil
}

const pluginTemplate = `//go:build {{.BuildTag}}
// +build {{.BuildTag}}

package {{.PackageName}}

import (
	"context"
	"mengri-flow/internal/infra/plugin"
)

// {{.TypeName}} 插件实现
type {{.TypeName}} struct{}

var _ plugin.{{.InterfaceName}} = (*{{.TypeName}})(nil)

func init() {
	plugin.GlobalRegistry().Register{{.InterfaceName}}(&{{.TypeName}}{})
}

// PluginMeta 返回插件元数据
func (p *{{.TypeName}}) PluginMeta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "{{.PluginName}}",
		Type:        plugin.PluginType{{.InterfaceName}},
		Version:     "1.0.0",
		Description: "{{.PluginName}} {{.PluginType}} plugin",
		Author:      "mengri-flow",
		BuildTag:    "{{.BuildTag}}",
	}
}

// ConfigSchema 返回配置Schema
func (p *{{.TypeName}}) ConfigSchema() plugin.JSONSchema {
	return plugin.BuildObjectSchema(
		"{{.PluginName}} Configuration",
		"Configuration for {{.PluginName}} {{.PluginType}}",
		map[string]plugin.JSONSchema{
			"endpoint": plugin.BuildStringSchema("Endpoint", "API endpoint", true),
		},
		[]string{"endpoint"},
	)
}
{{if eq .PluginType "resource"}}
// TestConnection 测试连接
func (p *{{.TypeName}}) TestConnection(ctx context.Context, config map[string]any) error {
	// TODO: 实现连接测试逻辑
	return nil
}

// ExecuteTool 执行工具
func (p *{{.TypeName}}) ExecuteTool(
	ctx context.Context,
	resourceConfig map[string]any,
	toolConfig map[string]any,
	input any,
) (*plugin.ToolResult, error) {
	// TODO: 实现工具执行逻辑
	return &plugin.ToolResult{
		StatusCode: 200,
		Data:       map[string]any{"result": "success"},
		Headers:    map[string]string{"Content-Type": "application/json"},
		Duration:   100,
		Error:      "",
	}, nil
}

// ExtractTools 提取工具定义
func (p *{{.TypeName}}) ExtractTools(ctx context.Context, config map[string]any) ([]plugin.ToolDefinition, error) {
	// TODO: 实现工具提取逻辑
	return []plugin.ToolDefinition{
		{
			Name:        "exampleTool",
			Type:        "http_method",
			Method:      "GET",
			Path:        "/example",
			Description: "Example tool",
			InputSchema: plugin.BuildObjectSchema(
				"Input",
				"Tool input",
				map[string]plugin.JSONSchema{
					"param1": plugin.BuildStringSchema("Parameter 1", "First parameter", true),
				},
				[]string{"param1"},
			),
			OutputSchema: plugin.BuildObjectSchema(
				"Output",
				"Tool output",
				map[string]plugin.JSONSchema{
					"result": plugin.BuildStringSchema("Result", "Execution result", false),
				},
				nil,
			),
		},
	}, nil
}
{{else}}
// InputSchema 返回输入Schema
func (p *{{.TypeName}}) InputSchema() plugin.JSONSchema {
	return plugin.BuildObjectSchema(
		"Trigger Input",
		"Input schema for trigger",
		map[string]plugin.JSONSchema{
			"data": plugin.BuildStringSchema("Data", "Input data", false),
		},
		nil,
	)
}

// OutputSchema 返回输出Schema
func (p *{{.TypeName}}) OutputSchema() plugin.JSONSchema {
	return plugin.BuildObjectSchema(
		"Trigger Output",
		"Output schema for trigger",
		map[string]plugin.JSONSchema{
			"result": plugin.BuildStringSchema("Result", "Trigger result", false),
		},
		nil,
	)
}

// Start 启动触发器
func (p *{{.TypeName}}) Start(ctx context.Context, config map[string]any, handler plugin.TriggerHandler) error {
	// TODO: 实现启动逻辑
	return nil
}

// Stop 停止触发器
func (p *{{.TypeName}}) Stop() error {
	// TODO: 实现停止逻辑
	return nil
}
{{end}}
`

const testTemplate = `//go:build {{.BuildTag}}
// +build {{.BuildTag}}

package {{.PackageName}}

import (
	"context"
	"testing"

	"mengri-flow/internal/infra/plugin"
)

func Test{{.TypeName}}_PluginMeta(t *testing.T) {
	p := &{{.TypeName}}{}
	meta := p.PluginMeta()

	if meta.Name != "{{.PluginName}}" {
		t.Errorf("expected name {{.PluginName}}, got %s", meta.Name)
	}

	if meta.Type != plugin.PluginType{{.InterfaceName}} {
		t.Errorf("expected type %s, got %s", plugin.PluginType{{.InterfaceName}}, meta.Type)
	}
}

func Test{{.TypeName}}_ConfigSchema(t *testing.T) {
	p := &{{.TypeName}}{}
	schema := p.ConfigSchema()

	if schema == nil {
		t.Fatal("ConfigSchema should not be nil")
	}

	// TODO: 添加更多Schema验证
}
{{if eq .PluginType "resource"}}
func Test{{.TypeName}}_TestConnection(t *testing.T) {
	p := &{{.TypeName}}{}
	ctx := context.Background()
	config := map[string]any{
		"endpoint": "http://example.com",
	}

	err := p.TestConnection(ctx, config)
	// TODO: 根据实际情况断言
	if err != nil {
		t.Logf("TestConnection returned error: %v", err)
	}
}

func Test{{.TypeName}}_ExecuteTool(t *testing.T) {
	p := &{{.TypeName}}{}
	ctx := context.Background()

	resourceConfig := map[string]any{
		"endpoint": "http://example.com",
	}
	toolConfig := map[string]any{}
	input := map[string]any{}

	result, err := p.ExecuteTool(ctx, resourceConfig, toolConfig, input)
	// TODO: 根据实际情况断言
	if err != nil {
		t.Logf("ExecuteTool returned error: %v", err)
	}
	if result == nil {
		t.Fatal("ExecuteTool should return a result")
	}
}
{{else}}
func Test{{.TypeName}}_StartStop(t *testing.T) {
	p := &{{.TypeName}}{}
	ctx := context.Background()

	config := map[string]any{}
	handler := func(ctx any, input map[string]any) (*plugin.TriggerResult, error) {
		return &plugin.TriggerResult{
			Success: true,
			Data:    map[string]any{"test": "data"},
		}, nil
	}

	err := p.Start(ctx, config, handler)
	// TODO: 根据实际情况断言
	if err != nil {
		t.Logf("Start returned error: %v", err)
	}

	err = p.Stop()
	// TODO: 根据实际情况断言
	if err != nil {
		t.Logf("Stop returned error: %v", err)
	}
}
{{end}}
`
