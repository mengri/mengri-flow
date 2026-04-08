package restful

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"mengri-flow/internal/infra/plugin"
)

func init() {
	registry := plugin.GlobalRegistry()
	registry.RegisterTrigger(&RESTfulTriggerPlugin{})
}

// 确保接口实现
var _ plugin.TriggerPlugin = (*RESTfulTriggerPlugin)(nil)

type RESTfulTriggerPlugin struct {
	mu       sync.RWMutex
	servers  map[string]*http.Server
	configs  map[string]map[string]interface{}
	handlers map[string]plugin.TriggerHandler
}

func (p *RESTfulTriggerPlugin) PluginMeta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "restful",
		Type:        plugin.PluginTypeTrigger,
		Version:     "1.0.0",
		Description: "RESTful Webhook触发器插件，支持同步和异步接口",
		Author:      "Platform Team",
	}
}

func (p *RESTfulTriggerPlugin) ConfigSchema() plugin.JSONSchema {
	return plugin.JSONSchema{
		"type": "object",
		"required": []string{"path", "method"},
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"title":       "请求路径",
				"description": "Webhook路径，如 /webhooks/order-created",
				"pattern":     "^/",
				"placeholder": "/webhooks/your-event",
			},
			"method": map[string]interface{}{
				"type":    "string",
				"title":   "HTTP方法",
				"enum":    []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
				"default": "POST",
			},
			"async": map[string]interface{}{
				"type":        "boolean",
				"title":       "异步模式",
				"description": "启用后返回202 Accepted，流程在后台执行",
				"default":     false,
			},
"port": map[string]interface{}{
			"type":        "integer",
			"title":       "监听端口",
			"description": "HTTP服务器监听端口（从etcd配置下发）",
			"minimum":     1024,
			"maximum":     65535,
			"readOnly":    true,
		},
		"auth": map[string]interface{}{
			"type":  "object",
			"title": "认证配置",
			"properties": map[string]interface{}{
				"type": map[string]interface{}{
					"type":    "string",
					"title":   "认证类型",
					"enum":    []string{"none", "apiKey"},
					"default": "none",
				},
				"apiKey": map[string]interface{}{
					"type":      "string",
					"title":     "API Key",
					"format":    "password",
					"condition": map[string]interface{}{"auth.type": "apiKey"},
				},
				"apiKeyLocation": map[string]interface{}{
					"type":      "string",
					"title":     "API Key位置",
					"enum":      []string{"header", "query"},
					"default":   "header",
					"condition": map[string]interface{}{"auth.type": "apiKey"},
				},
				"apiKeyName": map[string]interface{}{
					"type":      "string",
					"title":     "API Key名称",
					"default":   "X-API-Key",
					"condition": map[string]interface{}{"auth.type": "apiKey"},
				},
			},
		},
		},
	}
}

func (p *RESTfulTriggerPlugin) InputSchema() plugin.JSONSchema {
	return plugin.JSONSchema{
		"type": "object",
		"properties": map[string]interface{}{
			"headers": map[string]interface{}{
				"type":        "object",
				"description": "HTTP请求头",
			},
			"query": map[string]interface{}{
				"type":        "object",
				"description": "URL查询参数",
			},
			"body": map[string]interface{}{
				"type":        "object",
				"description": "请求体（JSON）",
			},
			"pathParams": map[string]interface{}{
				"type":        "object",
				"description": "路径参数",
			},
		},
	}
}

func (p *RESTfulTriggerPlugin) OutputSchema() plugin.JSONSchema {
	return plugin.JSONSchema{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":        "boolean",
				"description": "是否成功",
			},
			"data": map[string]interface{}{
				"type":        "object",
				"description": "响应数据",
			},
			"error": map[string]interface{}{
				"type":        "string",
				"description": "错误信息",
			},
		},
	}
}

func (p *RESTfulTriggerPlugin) Start(
	ctx context.Context,
	config map[string]interface{},
	handler plugin.TriggerHandler,
) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	triggerID, ok := config["triggerId"].(string)
	if !ok || triggerID == "" {
		return plugin.NewPluginError("config_error", "missing or invalid triggerId", nil)
	}

	if p.servers == nil {
		p.servers = make(map[string]*http.Server)
		p.configs = make(map[string]map[string]interface{})
		p.handlers = make(map[string]plugin.TriggerHandler)
	}

	p.configs[triggerID] = config
	p.handlers[triggerID] = handler

	path, ok := config["path"].(string)
	if !ok || path == "" {
		return plugin.NewPluginError("config_error", "missing or invalid path", nil)
	}

	method, ok := config["method"].(string)
	if !ok || method == "" {
		return plugin.NewPluginError("config_error", "missing or invalid method", nil)
	}

	mux := http.NewServeMux()

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "method not allowed",
			})
			return
		}

		if err := p.validateAuth(r, config); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "unauthorized",
			})
			return
		}

		input, err := p.parseRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("invalid request: %v", err),
			})
			return
		}

		async := false
		if val, ok := config["async"].(bool); ok {
			async = val
		}

		if async {
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"message": "request accepted, processing asynchronously",
			})

			go func() {
				p.executeInBackground(ctx, triggerID, input)
			}()
		} else {
			result, err := handler(ctx, input)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			if result.Success {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"data":    result.Data,
				})
			} else {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   result.Error,
				})
			}
		}
	})

	port, ok := config["port"].(int)
	if !ok || port == 0 {
		return plugin.NewPluginError("config_error", "missing or invalid port", nil)
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	p.servers[triggerID] = server

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("RESTful trigger server error: %v\n", err)
		}
	}()

	return nil
}

func (p *RESTfulTriggerPlugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for triggerID, server := range p.servers {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		if err := server.Shutdown(shutdownCtx); err != nil {
			cancel()
			return plugin.NewPluginError("internal_error", fmt.Sprintf("failed to shutdown server for trigger %s", triggerID), err)
		}

		cancel()
	}

	p.servers = nil
	p.configs = nil
	p.handlers = nil

	return nil
}

func (p *RESTfulTriggerPlugin) validateAuth(r *http.Request, config map[string]interface{}) error {
	authConfig, ok := config["auth"].(map[string]interface{})
	if !ok {
		return nil
	}

	authType, ok := authConfig["type"].(string)
	if !ok || authType == "none" {
		return nil
	}

	if authType == "apiKey" {
		apiKey, ok := authConfig["apiKey"].(string)
		if !ok || apiKey == "" {
			return fmt.Errorf("apiKey not configured")
		}

		location, _ := authConfig["apiKeyLocation"].(string)
		if location == "" {
			location = "header"
		}

		name, _ := authConfig["apiKeyName"].(string)
		if name == "" {
			name = "X-API-Key"
		}

		var providedKey string
		if location == "header" {
			providedKey = r.Header.Get(name)
		} else {
			providedKey = r.URL.Query().Get(name)
		}

		if providedKey != apiKey {
			return fmt.Errorf("invalid api key")
		}
	}

	return nil
}

func (p *RESTfulTriggerPlugin) parseRequest(r *http.Request) (map[string]interface{}, error) {
	input := make(map[string]interface{})

	headers := make(map[string]string)
	for key, values := range r.Header {
		headers[key] = strings.Join(values, ", ")
	}
	input["headers"] = headers

	query := make(map[string]string)
	for key, values := range r.URL.Query() {
		query[key] = strings.Join(values, ", ")
	}
	input["query"] = query

	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		defer r.Body.Close()

		if len(body) > 0 {
			var bodyJSON interface{}
			if err := json.Unmarshal(body, &bodyJSON); err != nil {
				input["body"] = string(body)
			} else {
				input["body"] = bodyJSON
			}
		}
	}

	input["pathParams"] = make(map[string]string)

	return input, nil
}

func (p *RESTfulTriggerPlugin) executeInBackground(ctx context.Context, triggerID string, input map[string]interface{}) {
	p.mu.RLock()
	handler, exists := p.handlers[triggerID]
	p.mu.RUnlock()

	if !exists {
		fmt.Printf("Handler not found for trigger %s\n", triggerID)
		return
	}

	_, err := handler(ctx, input)
	if err != nil {
		fmt.Printf("Background execution failed for trigger %s: %v\n", triggerID, err)
	}
}
