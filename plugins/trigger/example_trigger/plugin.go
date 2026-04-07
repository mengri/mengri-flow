package example_trigger

import (
	"context"
	"fmt"
	"sync"
	"time"

	"mengri-flow/internal/infra/plugin"
)

func init() {
	registry := plugin.GlobalRegistry()
	registry.RegisterTrigger(&ExampleTriggerPlugin{})
}

// ExampleTriggerPlugin 示例触发器插件
type ExampleTriggerPlugin struct {
	mu       sync.RWMutex
	handlers map[string]plugin.TriggerHandler
	configs  map[string]map[string]any
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
		"type":     "object",
		"required": []string{"interval"},
		"properties": map[string]any{
			"interval": map[string]any{
				"type":        "integer",
				"title":       "触发间隔（秒）",
				"description": "每隔多少秒触发一次",
				"minimum":     1,
				"maximum":     3600,
				"default":     60,
			},
		},
	}
}

// InputSchema 触发器输出数据 Schema
func (p *ExampleTriggerPlugin) InputSchema() plugin.JSONSchema {
	return plugin.JSONSchema{
		"type": "object",
		"properties": map[string]any{
			"timestamp": map[string]any{
				"type":        "string",
				"format":      "date-time",
				"description": "触发时间",
			},
			"triggerId": map[string]any{
				"type":        "string",
				"description": "触发器ID",
			},
		},
	}
}

// OutputSchema 响应 Schema（同步触发器用）
func (p *ExampleTriggerPlugin) OutputSchema() plugin.JSONSchema {
	return nil
}

// Start 启动触发器监听
func (p *ExampleTriggerPlugin) Start(ctx context.Context, config map[string]any, handler plugin.TriggerHandler) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	triggerID := config["triggerId"].(string)

	if p.handlers == nil {
		p.handlers = make(map[string]plugin.TriggerHandler)
		p.configs = make(map[string]map[string]any)
	}
	p.handlers[triggerID] = handler
	p.configs[triggerID] = config

	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel

	go p.run(triggerID, ctx, config, handler)

	return nil
}

func (p *ExampleTriggerPlugin) run(triggerID string, ctx context.Context, config map[string]any, handler plugin.TriggerHandler) {
	interval := time.Duration(config["interval"].(int)) * time.Second

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			input := map[string]any{
				"timestamp": t.Format(time.RFC3339),
				"triggerId": triggerID,
			}

			result, err := handler(ctx, input)
			if err != nil {
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
