package kafka

import (
	"context"
	"testing"

	"mengri-flow/internal/infra/plugin"
)

func TestKafkaTriggerPlugin_PluginMeta(t *testing.T) {
	p := &KafkaTriggerPlugin{}
	meta := p.PluginMeta()

	if meta.Name != "kafka" {
		t.Errorf("expected name 'kafka', got '%s'", meta.Name)
	}
	if meta.Type != "trigger" {
		t.Errorf("expected type 'trigger', got '%s'", meta.Type)
	}
	if meta.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", meta.Version)
	}
}

func TestKafkaTriggerPlugin_ConfigSchema(t *testing.T) {
	p := &KafkaTriggerPlugin{}
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

	requiredFields := []string{"brokers", "topic", "groupId", "autoCommit", "partition", "minBytes", "maxBytes", "maxWait"}
	for _, field := range requiredFields {
		if _, exists := properties[field]; !exists {
			t.Errorf("config schema missing field: %s", field)
		}
	}
}

func TestKafkaTriggerPlugin_InputOutputSchema(t *testing.T) {
	p := &KafkaTriggerPlugin{}

	inputSchema := p.InputSchema()
	if inputSchema["type"] != "object" {
		t.Errorf("expected input type 'object', got '%v'", inputSchema["type"])
	}

	inputProps, ok := inputSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("input properties should be a map")
	}

	requiredInputFields := []string{"key", "value", "topic", "partition", "offset", "timestamp"}
	for _, field := range requiredInputFields {
		if _, exists := inputProps[field]; !exists {
			t.Errorf("input schema missing field: %s", field)
		}
	}

	outputSchema := p.OutputSchema()
	if outputSchema != nil {
		t.Error("output schema should be nil for kafka trigger")
	}
}

func TestKafkaTriggerPlugin_InvalidConfig(t *testing.T) {
	p := &KafkaTriggerPlugin{}
	ctx := context.Background()

	tests := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name: "missing triggerId",
			config: map[string]interface{}{
				"brokers": []interface{}{"localhost:9092"},
				"topic":   "test-topic",
				"groupId": "test-group",
			},
		},
		{
			name: "missing brokers",
			config: map[string]interface{}{
				"triggerId": "test-kafka-1",
				"topic":     "test-topic",
				"groupId":   "test-group",
			},
		},
		{
			name: "empty brokers",
			config: map[string]interface{}{
				"triggerId": "test-kafka-2",
				"brokers":   []interface{}{},
				"topic":     "test-topic",
				"groupId":   "test-group",
			},
		},
		{
			name: "missing topic",
			config: map[string]interface{}{
				"triggerId": "test-kafka-3",
				"brokers":   []interface{}{"localhost:9092"},
				"groupId":   "test-group",
			},
		},
		{
			name: "missing groupId",
			config: map[string]interface{}{
				"triggerId": "test-kafka-4",
				"brokers":   []interface{}{"localhost:9092"},
				"topic":     "test-topic",
			},
		},
		{
			name: "empty triggerId",
			config: map[string]interface{}{
				"triggerId": "",
				"brokers":   []interface{}{"localhost:9092"},
				"topic":     "test-topic",
				"groupId":   "test-group",
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

func TestKafkaTriggerPlugin_BrokersFormat(t *testing.T) {
	p := &KafkaTriggerPlugin{}
	ctx := context.Background()

	config := map[string]interface{}{
		"triggerId": "test-kafka-5",
		"brokers":   []interface{}{"localhost:9092", "localhost:9093"},
		"topic":     "test-topic",
		"groupId":   "test-group",
	}

	var mockHandler plugin.TriggerHandler = func(ctx context.Context, input map[string]any) (*plugin.TriggerResult, error) {
		return &plugin.TriggerResult{
			Success: true,
			Data:    map[string]any{},
		}, nil
	}

	err := p.Start(ctx, config, mockHandler)
	if err != nil {
		t.Fatalf("failed to start plugin with multiple brokers: %v", err)
	}
	defer p.Stop()

	if len(p.readers) != 1 {
		t.Errorf("expected 1 reader, got %d", len(p.readers))
	}
}

func TestKafkaTriggerPlugin_PartitionConfig(t *testing.T) {
	p := &KafkaTriggerPlugin{}
	ctx := context.Background()

	config := map[string]interface{}{
		"triggerId": "test-kafka-6",
		"brokers":   []interface{}{"localhost:9092"},
		"topic":     "test-topic",
		"partition": 0,
	}

	var mockHandler plugin.TriggerHandler = func(ctx context.Context, input map[string]any) (*plugin.TriggerResult, error) {
		return &plugin.TriggerResult{
			Success: true,
			Data:    map[string]any{},
		}, nil
	}

	err := p.Start(ctx, config, mockHandler)
	if err != nil {
		t.Fatalf("failed to start plugin with partition config: %v", err)
	}
	defer p.Stop()
}

func TestKafkaTriggerPlugin_Stop(t *testing.T) {
	p := &KafkaTriggerPlugin{}

	p.readers = map[string]*kafka.Reader{
		"test-trigger-1": nil,
		"test-trigger-2": nil,
	}
	p.configs = map[string]map[string]interface{}{
		"test-trigger-1": {},
		"test-trigger-2": {},
	}
	p.handlers = map[string]plugin.TriggerHandler{}

	p.Stop()

	if p.readers != nil {
		t.Error("readers map should be nil after Stop")
	}
	if p.configs != nil {
		t.Error("configs map should be nil after Stop")
	}
	if p.handlers != nil {
		t.Error("handlers map should be nil after Stop")
	}
}

func TestKafkaTriggerPlugin_ConfigDefaults(t *testing.T) {
	p := &KafkaTriggerPlugin{}
	schema := p.ConfigSchema()

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties should be a map")
	}

	autoCommit, ok := properties["autoCommit"].(map[string]interface{})
	if !ok {
		t.Fatal("autoCommit should be a map")
	}
	if autoCommit["default"] != true {
		t.Errorf("expected default autoCommit true, got %v", autoCommit["default"])
	}

	minBytes, ok := properties["minBytes"].(map[string]interface{})
	if !ok {
		t.Fatal("minBytes should be a map")
	}
	if minBytes["default"] != 10240 {
		t.Errorf("expected default minBytes 10240, got %v", minBytes["default"])
	}

	maxBytes, ok := properties["maxBytes"].(map[string]interface{})
	if !ok {
		t.Fatal("maxBytes should be a map")
	}
	if maxBytes["default"] != 10485760 {
		t.Errorf("expected default maxBytes 10485760, got %v", maxBytes["default"])
	}

	maxWait, ok := properties["maxWait"].(map[string]interface{})
	if !ok {
		t.Fatal("maxWait should be a map")
	}
	if maxWait["default"] != 1000 {
		t.Errorf("expected default maxWait 1000, got %v", maxWait["default"])
	}
}
