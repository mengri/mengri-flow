package timer

import (
	"context"
	"sync"
	"testing"
	"time"

	"mengri-flow/internal/infra/plugin"
)

func TestTimerTriggerPlugin_PluginMeta(t *testing.T) {
	p := &TimerTriggerPlugin{}
	meta := p.PluginMeta()

	if meta.Name != "timer" {
		t.Errorf("expected name 'timer', got '%s'", meta.Name)
	}
	if meta.Type != "trigger" {
		t.Errorf("expected type 'trigger', got '%s'", meta.Type)
	}
	if meta.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", meta.Version)
	}
}

func TestTimerTriggerPlugin_ConfigSchema(t *testing.T) {
	p := &TimerTriggerPlugin{}
	schema := p.ConfigSchema()

	if schema["type"] != "object" {
		t.Errorf("expected type 'object', got '%v'", schema["type"])
	}

	required, ok := schema["required"].([]string)
	if !ok || len(required) != 1 || required[0] != "cronExpression" {
		t.Errorf("expected required field 'cronExpression', got %v", required)
	}

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties should be a map")
	}

	requiredFields := []string{"cronExpression", "timezone", "distributedLock", "lockKey", "lockTTL"}
	for _, field := range requiredFields {
		if _, exists := properties[field]; !exists {
			t.Errorf("config schema missing field: %s", field)
		}
	}
}

func TestTimerTriggerPlugin_InputOutputSchema(t *testing.T) {
	p := &TimerTriggerPlugin{}

	inputSchema := p.InputSchema()
	if inputSchema["type"] != "object" {
		t.Errorf("expected input type 'object', got '%v'", inputSchema["type"])
	}

	inputProps, ok := inputSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("input properties should be a map")
	}

	requiredInputFields := []string{"triggerTime", "executionId", "triggerId"}
	for _, field := range requiredInputFields {
		if _, exists := inputProps[field]; !exists {
			t.Errorf("input schema missing field: %s", field)
		}
	}

	outputSchema := p.OutputSchema()
	if outputSchema != nil {
		t.Error("output schema should be nil for timer trigger")
	}
}

func TestTimerTriggerPlugin_Start_Stop(t *testing.T) {
	p := &TimerTriggerPlugin{}
	ctx := context.Background()

	config := map[string]interface{}{
		"triggerId":      "test-timer-1",
		"cronExpression": "*/1 * * * * *", // 每秒执行一次（测试用）
		"timezone":       "Asia/Shanghai",
	}

	executionCount := 0
	var mockHandler plugin.TriggerHandler = func(ctx context.Context, input map[string]any) (*plugin.TriggerResult, error) {
		executionCount++
		return &plugin.TriggerResult{
			Success: true,
			Data:    map[string]any{"count": executionCount},
		}, nil
	}

	err := p.Start(ctx, config, mockHandler)
	if err != nil {
		t.Fatalf("failed to start plugin: %v", err)
	}

	if len(p.crons) != 1 {
		t.Errorf("expected 1 cron instance, got %d", len(p.crons))
	}

	time.Sleep(1500 * time.Millisecond)

	if executionCount < 1 {
		t.Errorf("expected at least 1 execution, got %d", executionCount)
	}

	p.Stop()

	if p.crons != nil {
		t.Error("crons map should be nil after Stop")
	}
}

func TestTimerTriggerPlugin_CronExecution(t *testing.T) {
	p := &TimerTriggerPlugin{}
	ctx := context.Background()

	config := map[string]interface{}{
		"triggerId":      "test-timer-2",
		"cronExpression": "*/1 * * * * *", // 每秒执行一次
		"timezone":       "UTC",
	}

	executionDetails := make([]map[string]any, 0)
	var mockHandler plugin.TriggerHandler = func(ctx context.Context, input map[string]any) (*plugin.TriggerResult, error) {
		executionDetails = append(executionDetails, input)
		return &plugin.TriggerResult{
			Success: true,
			Data:    map[string]any{"status": "completed"},
		}, nil
	}

	err := p.Start(ctx, config, mockHandler)
	if err != nil {
		t.Fatalf("failed to start plugin: %v", err)
	}
	defer p.Stop()

	time.Sleep(2500 * time.Millisecond)

	if len(executionDetails) < 2 {
		t.Errorf("expected at least 2 executions, got %d", len(executionDetails))
	}

	for _, detail := range executionDetails {
		if _, ok := detail["triggerTime"].(string); !ok {
			t.Error("triggerTime should be a string")
		}
		if executionID, ok := detail["executionId"].(string); !ok || executionID == "" {
			t.Error("executionId should be a non-empty string")
		}
		if triggerID, ok := detail["triggerId"].(string); !ok || triggerID != "test-timer-2" {
			t.Errorf("triggerId should be 'test-timer-2', got %v", triggerID)
		}
	}
}

func TestTimerTriggerPlugin_InvalidCronExpression(t *testing.T) {
	p := &TimerTriggerPlugin{}
	ctx := context.Background()

	config := map[string]interface{}{
		"triggerId":      "test-timer-3",
		"cronExpression": "invalid-cron",
		"timezone":       "Asia/Shanghai",
	}

	var mockHandler plugin.TriggerHandler = func(ctx context.Context, input map[string]any) (*plugin.TriggerResult, error) {
		return &plugin.TriggerResult{
			Success: true,
			Data:    map[string]any{},
		}, nil
	}

	err := p.Start(ctx, config, mockHandler)
	if err == nil {
		t.Error("expected error for invalid cron expression, got nil")
	}
}

func TestTimerTriggerPlugin_InvalidTimezone(t *testing.T) {
	p := &TimerTriggerPlugin{}
	ctx := context.Background()

	config := map[string]interface{}{
		"triggerId":      "test-timer-4",
		"cronExpression": "0 2 * * *",
		"timezone":       "Invalid/Timezone",
	}

	var mockHandler plugin.TriggerHandler = func(ctx context.Context, input map[string]any) (*plugin.TriggerResult, error) {
		return &plugin.TriggerResult{
			Success: true,
			Data:    map[string]any{},
		}, nil
	}

	err := p.Start(ctx, config, mockHandler)
	if err == nil {
		t.Error("expected error for invalid timezone, got nil")
	}
}

func TestTimerTriggerPlugin_MissingRequiredFields(t *testing.T) {
	p := &TimerTriggerPlugin{}
	ctx := context.Background()

	tests := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name: "missing triggerId",
			config: map[string]interface{}{
				"cronExpression": "0 2 * * *",
			},
		},
		{
			name: "missing cronExpression",
			config: map[string]interface{}{
				"triggerId": "test-timer-5",
			},
		},
		{
			name: "empty triggerId",
			config: map[string]interface{}{
				"triggerId":      "",
				"cronExpression": "0 2 * * *",
			},
		},
		{
			name: "empty cronExpression",
			config: map[string]interface{}{
				"triggerId":      "test-timer-6",
				"cronExpression": "",
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

func TestTimerTriggerPlugin_ConcurrentExecutions(t *testing.T) {
	p := &TimerTriggerPlugin{}
	ctx := context.Background()

	config := map[string]interface{}{
		"triggerId":      "test-timer-7",
		"cronExpression": "*/1 * * * * *", // 每秒执行一次
		"timezone":       "UTC",
	}

	executionCount := 0
	var mu sync.Mutex
	var mockHandler plugin.TriggerHandler = func(ctx context.Context, input map[string]any) (*plugin.TriggerResult, error) {
		mu.Lock()
		executionCount++
		mu.Unlock()
		time.Sleep(100 * time.Millisecond) // 模拟耗时操作
		return &plugin.TriggerResult{
			Success: true,
			Data:    map[string]any{},
		}, nil
	}

	err := p.Start(ctx, config, mockHandler)
	if err != nil {
		t.Fatalf("failed to start plugin: %v", err)
	}
	defer p.Stop()

	time.Sleep(2200 * time.Millisecond)

	mu.Lock()
	count := executionCount
	mu.Unlock()

	if count < 2 {
		t.Errorf("expected at least 2 executions, got %d", count)
	}
}

func TestGenerateExecutionID(t *testing.T) {
	id1 := generateExecutionID()
	id2 := generateExecutionID()

	if id1 == id2 {
		t.Error("execution IDs should be unique")
	}

	if len(id1) == 0 || len(id2) == 0 {
		t.Error("execution IDs should not be empty")
	}

	if !contains(id1, "timer-") || !contains(id2, "timer-") {
		t.Error("execution IDs should start with 'timer-'")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
