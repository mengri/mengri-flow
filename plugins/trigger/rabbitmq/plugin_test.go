package rabbitmq

import (
	"context"
	"testing"

	"mengri-flow/internal/infra/plugin"
)

func TestRabbitMQTriggerPlugin_PluginMeta(t *testing.T) {
	p := &RabbitMQTriggerPlugin{}
	meta := p.PluginMeta()

	if meta.Name != "rabbitmq" {
		t.Errorf("expected name 'rabbitmq', got '%s'", meta.Name)
	}
	if meta.Type != "trigger" {
		t.Errorf("expected type 'trigger', got '%s'", meta.Type)
	}
	if meta.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", meta.Version)
	}
}

func TestRabbitMQTriggerPlugin_ConfigSchema(t *testing.T) {
	p := &RabbitMQTriggerPlugin{}
	schema := p.ConfigSchema()

	if schema["type"] != "object" {
		t.Errorf("expected type 'object', got '%v'", schema["type"])
	}

	required, ok := schema["required"].([]string)
	if !ok || len(required) != 3 {
		t.Errorf("expected 3 required fields, got %v", required)
	}

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties should be a map")
	}

	requiredFields := []string{"brokerUrl", "queue", "consumerTag", "autoAck", "prefetchCount"}
	for _, field := range requiredFields {
		if _, exists := properties[field]; !exists {
			t.Errorf("config schema missing field: %s", field)
		}
	}
}

func TestRabbitMQTriggerPlugin_InputOutputSchema(t *testing.T) {
	p := &RabbitMQTriggerPlugin{}

	inputSchema := p.InputSchema()
	if inputSchema["type"] != "object" {
		t.Errorf("expected input type 'object', got '%v'", inputSchema["type"])
	}

	inputProps, ok := inputSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("input properties should be a map")
	}

	requiredInputFields := []string{"messageId", "body", "headers", "timestamp", "deliveryTag"}
	for _, field := range requiredInputFields {
		if _, exists := inputProps[field]; !exists {
			t.Errorf("input schema missing field: %s", field)
		}
	}

	outputSchema := p.OutputSchema()
	if outputSchema != nil {
		t.Error("output schema should be nil for rabbitmq trigger")
	}
}

func TestRabbitMQTriggerPlugin_InvalidConfig(t *testing.T) {
	p := &RabbitMQTriggerPlugin{}
	ctx := context.Background()

	tests := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name: "missing triggerId",
			config: map[string]interface{}{
				"brokerUrl":     "amqp://localhost:5672",
				"queue":         "test-queue",
				"consumerTag":   "test-consumer",
			},
		},
		{
			name: "missing brokerUrl",
			config: map[string]interface{}{
				"triggerId":   "test-rabbit-1",
				"queue":       "test-queue",
				"consumerTag": "test-consumer",
			},
		},
		{
			name: "missing queue",
			config: map[string]interface{}{
				"triggerId":   "test-rabbit-2",
				"brokerUrl":   "amqp://localhost:5672",
				"consumerTag": "test-consumer",
			},
		},
		{
			name: "missing consumerTag",
			config: map[string]interface{}{
				"triggerId": "test-rabbit-3",
				"brokerUrl": "amqp://localhost:5672",
				"queue":     "test-queue",
			},
		},
		{
			name: "empty triggerId",
			config: map[string]interface{}{
				"triggerId":     "",
				"brokerUrl":     "amqp://localhost:5672",
				"queue":         "test-queue",
				"consumerTag":   "test-consumer",
			},
		},
	}

	var mockHandler plugin.TriggerHandler = func(ctx context.Context, input map[string]any) (*plugin.TriggerResult, error) {
		return &plugin.TriggerResult{
			Success: true,
			Data:    map[string]any{},
		}, nil
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.Start(ctx, tt.config, mockHandler)
			if err == nil {
				t.Errorf("expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestRabbitMQTriggerPlugin_Stop(t *testing.T) {
	p := &RabbitMQTriggerPlugin{}

	p.channels = map[string]*amqp.Channel{
		"test-trigger-1": nil,
		"test-trigger-2": nil,
	}
	p.configs = map[string]map[string]interface{}{
		"test-trigger-1": {},
		"test-trigger-2": {},
	}
	p.handlers = map[string]plugin.TriggerHandler{}

	p.Stop()

	if p.channels != nil {
		t.Error("channels map should be nil after Stop")
	}
	if p.configs != nil {
		t.Error("configs map should be nil after Stop")
	}
	if p.handlers != nil {
		t.Error("handlers map should be nil after Stop")
	}
}

func TestRabbitMQTriggerPlugin_PrefetchCount(t *testing.T) {
	p := &RabbitMQTriggerPlugin{}
	schema := p.ConfigSchema()

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties should be a map")
	}

	prefetchCount, ok := properties["prefetchCount"].(map[string]interface{})
	if !ok {
		t.Fatal("prefetchCount should be a map")
	}

	if prefetchCount["default"] != 10 {
		t.Errorf("expected default prefetchCount 10, got %v", prefetchCount["default"])
	}
	if prefetchCount["minimum"] != 1 {
		t.Errorf("expected minimum prefetchCount 1, got %v", prefetchCount["minimum"])
	}
	if prefetchCount["maximum"] != 1000 {
		t.Errorf("expected maximum prefetchCount 1000, got %v", prefetchCount["maximum"])
	}
}

func TestRabbitMQTriggerPlugin_AutoAckDefault(t *testing.T) {
	p := &RabbitMQTriggerPlugin{}
	schema := p.ConfigSchema()

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties should be a map")
	}

	autoAck, ok := properties["autoAck"].(map[string]interface{})
	if !ok {
		t.Fatal("autoAck should be a map")
	}

	if autoAck["default"] != false {
		t.Errorf("expected default autoAck false, got %v", autoAck["default"])
	}
}
