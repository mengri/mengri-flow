package restful

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mengri-flow/internal/infra/plugin"
)

func TestRESTfulTriggerPlugin_PluginMeta(t *testing.T) {
	p := &RESTfulTriggerPlugin{}
	meta := p.PluginMeta()

	if meta.Name != "restful" {
		t.Errorf("expected name 'restful', got '%s'", meta.Name)
	}
	if meta.Type != "trigger" {
		t.Errorf("expected type 'trigger', got '%s'", meta.Type)
	}
	if meta.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", meta.Version)
	}
}

func TestRESTfulTriggerPlugin_ConfigSchema(t *testing.T) {
	p := &RESTfulTriggerPlugin{}
	schema := p.ConfigSchema()

	if schema["type"] != "object" {
		t.Errorf("expected type 'object', got '%v'", schema["type"])
	}

	required, ok := schema["required"].([]string)
	if !ok || len(required) != 2 {
		t.Errorf("expected 2 required fields, got %v", required)
	}

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties should be a map")
	}

	if _, exists := properties["path"]; !exists {
		t.Error("path property is required")
	}
	if _, exists := properties["method"]; !exists {
		t.Error("method property is required")
	}
	if _, exists := properties["async"]; !exists {
		t.Error("async property is required")
	}
	if _, exists := properties["auth"]; !exists {
		t.Error("auth property is required")
	}
}

func TestRESTfulTriggerPlugin_InputOutputSchema(t *testing.T) {
	p := &RESTfulTriggerPlugin{}

	inputSchema := p.InputSchema()
	if inputSchema["type"] != "object" {
		t.Errorf("expected input type 'object', got '%v'", inputSchema["type"])
	}

	inputProps, ok := inputSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("input properties should be a map")
	}

	requiredFields := []string{"headers", "query", "body", "pathParams"}
	for _, field := range requiredFields {
		if _, exists := inputProps[field]; !exists {
			t.Errorf("input schema missing field: %s", field)
		}
	}

	outputSchema := p.OutputSchema()
	if outputSchema["type"] != "object" {
		t.Errorf("expected output type 'object', got '%v'", outputSchema["type"])
	}

	outputProps, ok := outputSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("output properties should be a map")
	}

	if _, exists := outputProps["success"]; !exists {
		t.Error("output schema missing 'success' field")
	}
	if _, exists := outputProps["data"]; !exists {
		t.Error("output schema missing 'data' field")
	}
	if _, exists := outputProps["error"]; !exists {
		t.Error("output schema missing 'error' field")
	}
}

func TestRESTfulTriggerPlugin_SyncExecution(t *testing.T) {
	p := &RESTfulTriggerPlugin{}
	ctx := context.Background()

	config := map[string]interface{}{
		"triggerId": "test-trigger-1",
		"path":      "/webhooks/test",
		"method":    "POST",
		"port":      18081,
		"async":     false,
	}

	var mockHandler plugin.TriggerHandler = func(ctx any, input map[string]any) (*plugin.TriggerResult, error) {
		return &plugin.TriggerResult{
			Success: true,
			Data:    map[string]any{"message": "processed"},
		}, nil
	}

	err := p.Start(ctx, config, mockHandler)
	if err != nil {
		t.Fatalf("failed to start plugin: %v", err)
	}
	defer p.Stop()

	payload := map[string]string{"event": "test"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/webhooks/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	server := p.servers["test-trigger-1"]
	if server == nil {
		t.Fatal("server not found")
	}

	server.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Errorf("expected success=true, got %v", response["success"])
	}

	if data, ok := response["data"].(map[string]interface{}); !ok {
		t.Error("response data should be an object")
	} else if msg, ok := data["message"].(string); !ok || msg != "processed" {
		t.Errorf("expected message 'processed', got %v", data["message"])
	}
}

func TestRESTfulTriggerPlugin_AsyncExecution(t *testing.T) {
	p := &RESTfulTriggerPlugin{}
	ctx := context.Background()

	config := map[string]interface{}{
		"triggerId": "test-trigger-2",
		"path":      "/webhooks/test-async",
		"method":    "POST",
		"port":      18082,
		"async":     true,
	}

	executionCount := 0
	var mockHandler plugin.TriggerHandler = func(ctx any, input map[string]any) (*plugin.TriggerResult, error) {
		executionCount++
		return &plugin.TriggerResult{
			Success: true,
			Data:    map[string]any{"status": "background"},
		}, nil
	}

	err := p.Start(ctx, config, mockHandler)
	if err != nil {
		t.Fatalf("failed to start plugin: %v", err)
	}
	defer p.Stop()

	payload := map[string]string{"event": "async-test"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/webhooks/test-async", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	server := p.servers["test-trigger-2"]
	if server == nil {
		t.Fatal("server not found")
	}

	server.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Errorf("expected success=true, got %v", response["success"])
	}

	time.Sleep(100 * time.Millisecond)

	if executionCount != 1 {
		t.Errorf("expected handler to be executed once, got %d", executionCount)
	}
}

func TestRESTfulTriggerPlugin_AuthValidation(t *testing.T) {
	tests := []struct {
		name           string
		config         map[string]interface{}
		requestHeaders map[string]string
		requestQuery   map[string]string
		expectStatus   int
	}{
		{
			name: "no_auth_required",
			config: map[string]interface{}{
				"triggerId": "test-auth-1",
				"path":      "/webhooks/noauth",
				"method":    "POST",
				"port":      18083,
				"async":     false,
				"auth": map[string]interface{}{
					"type": "none",
				},
			},
			requestHeaders: map[string]string{},
			expectStatus:   http.StatusOK,
		},
		{
			name: "apiKey_header_success",
			config: map[string]interface{}{
				"triggerId": "test-auth-2",
				"path":      "/webhooks/auth-header",
				"method":    "POST",
				"port":      18084,
				"async":     false,
				"auth": map[string]interface{}{
					"type":           "apiKey",
					"apiKey":         "secret-key-123",
					"apiKeyLocation": "header",
					"apiKeyName":     "X-API-Key",
				},
			},
			requestHeaders: map[string]string{
				"X-API-Key": "secret-key-123",
			},
			expectStatus: http.StatusOK,
		},
		{
			name: "apiKey_header_failure",
			config: map[string]interface{}{
				"triggerId": "test-auth-3",
				"path":      "/webhooks/auth-header-fail",
				"method":    "POST",
				"port":      18085,
				"async":     false,
				"auth": map[string]interface{}{
					"type":           "apiKey",
					"apiKey":         "secret-key-123",
					"apiKeyLocation": "header",
					"apiKeyName":     "X-API-Key",
				},
			},
			requestHeaders: map[string]string{
				"X-API-Key": "wrong-key",
			},
			expectStatus: http.StatusUnauthorized,
		},
		{
			name: "apiKey_query_success",
			config: map[string]interface{}{
				"triggerId": "test-auth-4",
				"path":      "/webhooks/auth-query",
				"method":    "POST",
				"port":      18086,
				"async":     false,
				"auth": map[string]interface{}{
					"type":           "apiKey",
					"apiKey":         "secret-key-456",
					"apiKeyLocation": "query",
					"apiKeyName":     "api_key",
				},
			},
			requestQuery: map[string]string{
				"api_key": "secret-key-456",
			},
			expectStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &RESTfulTriggerPlugin{}
			ctx := context.Background()

			var mockHandler plugin.TriggerHandler = func(ctx any, input map[string]any) (*plugin.TriggerResult, error) {
				return &plugin.TriggerResult{
					Success: true,
					Data:    map[string]any{"status": "ok"},
				}, nil
			}

triggerID := tt.config["triggerId"].(string)
		err := p.Start(ctx, tt.config, mockHandler)
		if err != nil {
			t.Fatalf("failed to start plugin: %v", err)
		}
		defer p.Stop()

			body := bytes.NewReader([]byte("{}"))
			req := httptest.NewRequest("POST", tt.config["path"].(string), body)

			for key, value := range tt.requestHeaders {
				req.Header.Set(key, value)
			}

			query := req.URL.Query()
			for key, value := range tt.requestQuery {
				query.Add(key, value)
			}
			req.URL.RawQuery = query.Encode()

			w := httptest.NewRecorder()

			server := p.servers[triggerID]
			if server == nil {
				t.Fatal("server not found")
			}

			server.Handler.ServeHTTP(w, req)

			if w.Code != tt.expectStatus {
				t.Errorf("expected status %d, got %d", tt.expectStatus, w.Code)
			}
		})
	}
}

func TestRESTfulTriggerPlugin_MethodValidation(t *testing.T) {
	p := &RESTfulTriggerPlugin{}
	ctx := context.Background()

	config := map[string]interface{}{
		"triggerId": "test-method",
		"path":      "/webhooks/method-test",
		"method":    "POST",
		"port":      18087,
		"async":     false,
	}

	var mockHandler plugin.TriggerHandler = func(ctx any, input map[string]any) (*plugin.TriggerResult, error) {
		return &plugin.TriggerResult{
			Success: true,
			Data:    map[string]any{"status": "ok"},
		}, nil
	}

	err := p.Start(ctx, config, mockHandler)
	if err != nil {
		t.Fatalf("failed to start plugin: %v", err)
	}
	defer p.Stop()

	tests := []struct {
		method       string
		expectStatus int
	}{
		{"POST", http.StatusOK},
		{"GET", http.StatusMethodNotAllowed},
		{"PUT", http.StatusMethodNotAllowed},
		{"DELETE", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/webhooks/method-test", bytes.NewReader([]byte("{}")))
			w := httptest.NewRecorder()

			server := p.servers["test-method"]
			server.Handler.ServeHTTP(w, req)

			if w.Code != tt.expectStatus {
				t.Errorf("for method %s: expected status %d, got %d", tt.method, tt.expectStatus, w.Code)
			}
		})
	}
}

func TestRESTfulTriggerPlugin_Stop(t *testing.T) {
	p := &RESTfulTriggerPlugin{}
	ctx := context.Background()

	config := map[string]interface{}{
		"triggerId": "test-stop",
		"path":      "/webhooks/stop-test",
		"method":    "POST",
		"port":      18088,
		"async":     false,
	}

	var mockHandler plugin.TriggerHandler = func(ctx any, input map[string]any) (*plugin.TriggerResult, error) {
		return &plugin.TriggerResult{
			Success: true,
			Data:    map[string]any{"status": "ok"},
		}, nil
	}

	err := p.Start(ctx, config, mockHandler)
	if err != nil {
		t.Fatalf("failed to start plugin: %v", err)
	}

	if _, exists := p.servers["test-stop"]; !exists {
		t.Error("server should exist after Start")
	}

	err = p.Stop()
	if err != nil {
		t.Errorf("failed to stop plugin: %v", err)
	}

	if _, exists := p.servers["test-stop"]; exists {
		t.Error("server should not exist after Stop")
	}
}

func TestRESTfulTriggerPlugin_ConcurrentRequests(t *testing.T) {
	p := &RESTfulTriggerPlugin{}
	ctx := context.Background()

	config := map[string]interface{}{
		"triggerId": "test-concurrent",
		"path":      "/webhooks/concurrent",
		"method":    "POST",
		"port":      18089,
		"async":     false,
	}

	requestCount := 0
	var mockHandler plugin.TriggerHandler = func(ctx any, input map[string]any) (*plugin.TriggerResult, error) {
		requestCount++
		time.Sleep(10 * time.Millisecond)
		return &plugin.TriggerResult{
			Success: true,
			Data:    map[string]any{"count": requestCount},
		}, nil
	}

	err := p.Start(ctx, config, mockHandler)
	if err != nil {
		t.Fatalf("failed to start plugin: %v", err)
	}
	defer p.Stop()

	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			payload := map[string]string{"event": "concurrent-test"}
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/webhooks/concurrent", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			server := p.servers["test-concurrent"]
			server.Handler.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				done <- true
			} else {
				done <- false
			}
		}()
	}

	successCount := 0
	for i := 0; i < concurrency; i++ {
		if <-done {
			successCount++
		}
	}

	if successCount != concurrency {
		t.Errorf("expected all %d requests to succeed, got %d", concurrency, successCount)
	}

	if requestCount != concurrency {
		t.Errorf("expected handler to be called %d times, got %d", concurrency, requestCount)
	}
}
