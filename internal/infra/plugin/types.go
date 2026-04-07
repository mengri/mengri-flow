package plugin

import "context"

// PluginType 插件类型
type PluginType string

const (
	PluginTypeResource PluginType = "resource"
	PluginTypeTrigger  PluginType = "trigger"
)

// PluginMeta 插件元数据
type PluginMeta struct {
	Name        string     // 插件名称
	Type        PluginType // 插件类型
	Version     string     // 插件版本
	Description string     // 插件描述
	Author      string     // 插件作者
	BuildTag    string     // 构建标签
}

// JSONSchema JSON Schema 定义
type JSONSchema map[string]any

// ToolDefinition 工具定义
type ToolDefinition struct {
	Name         string         // 工具名称
	Type         string         // 工具类型
	Method       string         // HTTP 方法
	Path         string         // 请求路径
	Description  string         // 工具描述
	InputSchema  JSONSchema     // 输入 Schema
	OutputSchema JSONSchema     // 输出 Schema
	Config       map[string]any // 工具配置
}

// ToolResult 工具执行结果
type ToolResult struct {
	StatusCode int               // HTTP 状态码
	Data       any               // 返回数据
	Headers    map[string]string // 响应头
	Duration   any               // 执行耗时
	Error      string            // 错误信息
}

// TriggerHandler 触发器回调函数
type TriggerHandler func(ctx any, input map[string]any) (*TriggerResult, error)

// TriggerResult 触发处理结果
type TriggerResult struct {
	Success bool           // 是否成功触发
	Data    map[string]any // 响应数据
	Error   string         // 错误信息
}

// ResourcePlugin 资源插件接口
type ResourcePlugin interface {
	PluginMeta() PluginMeta
	ConfigSchema() JSONSchema
	TestConnection(ctx context.Context, config map[string]any) error
	ExecuteTool(
		ctx context.Context,
		resourceConfig map[string]any,
		toolConfig map[string]any,
		input any,
	) (*ToolResult, error)
	ExtractTools(ctx context.Context, config map[string]any) ([]ToolDefinition, error)
}

// TriggerPlugin 触发器插件接口
type TriggerPlugin interface {
	PluginMeta() PluginMeta
	ConfigSchema() JSONSchema
	InputSchema() JSONSchema
	OutputSchema() JSONSchema
	Start(ctx context.Context, config map[string]any, handler TriggerHandler) error
	Stop() error
}
