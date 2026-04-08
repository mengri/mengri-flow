package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"mengri-flow/internal/infra/plugin"
)

func init() {
	registry := plugin.GlobalRegistry()
	registry.RegisterTrigger(&RabbitMQTriggerPlugin{})
}

// 确保接口实现
var _ plugin.TriggerPlugin = (*RabbitMQTriggerPlugin)(nil)

type RabbitMQTriggerPlugin struct {
	mu       sync.RWMutex
	channels map[string]*amqp.Channel
	configs  map[string]map[string]interface{}
	handlers map[string]plugin.TriggerHandler
	conn     *amqp.Connection
}

func (p *RabbitMQTriggerPlugin) PluginMeta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "rabbitmq",
		Type:        plugin.PluginTypeTrigger,
		Version:     "1.0.0",
		Description: "RabbitMQ消息触发器插件",
		Author:      "Platform Team",
	}
}

func (p *RabbitMQTriggerPlugin) ConfigSchema() plugin.JSONSchema {
	return plugin.JSONSchema{
		"type": "object",
		"required": []string{"brokerUrl", "queue", "consumerTag"},
		"properties": map[string]interface{}{
			"brokerUrl": map[string]interface{}{
				"type":        "string",
				"title":       "Broker URL",
				"description": "RabbitMQ连接地址，如 amqp://guest:guest@localhost:5672/",
				"format":      "uri",
				"placeholder": "amqp://user:pass@localhost:5672/",
			},
			"queue": map[string]interface{}{
				"type":        "string",
				"title":       "队列名称",
				"description": "监听的队列名",
				"placeholder": "my-queue",
			},
			"consumerTag": map[string]interface{}{
				"type":        "string",
				"title":       "消费者标签",
				"description": "消费者唯一标识",
				"placeholder": "consumer-001",
			},
			"autoAck": map[string]interface{}{
				"type":        "boolean",
				"title":       "自动确认",
				"description": "收到消息后自动确认",
				"default":     false,
			},
			"prefetchCount": map[string]interface{}{
				"type":        "integer",
				"title":       "预取数量",
				"description": "一次预取的消息数量",
				"default":     10,
				"minimum":     1,
				"maximum":     1000,
			},
		},
	}
}

func (p *RabbitMQTriggerPlugin) InputSchema() plugin.JSONSchema {
	return plugin.JSONSchema{
		"type": "object",
		"properties": map[string]interface{}{
			"messageId": map[string]interface{}{
				"type":        "string",
				"description": "消息ID",
			},
			"body": map[string]interface{}{
				"type":        "object",
				"description": "消息体（JSON解析后）",
			},
			"headers": map[string]interface{}{
				"type":        "object",
				"description": "消息头",
			},
			"timestamp": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "消息时间戳",
			},
			"deliveryTag": map[string]interface{}{
				"type":        "integer",
				"description": "消息投递标签",
			},
		},
	}
}

func (p *RabbitMQTriggerPlugin) OutputSchema() plugin.JSONSchema {
	return nil
}

func (p *RabbitMQTriggerPlugin) Start(
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

	if p.channels == nil {
		p.channels = make(map[string]*amqp.Channel)
		p.configs = make(map[string]map[string]interface{})
		p.handlers = make(map[string]plugin.TriggerHandler)
	}

	p.configs[triggerID] = config
	p.handlers[triggerID] = handler

	brokerUrl, ok := config["brokerUrl"].(string)
	if !ok || brokerUrl == "" {
		return plugin.NewPluginError("config_error", "missing or invalid brokerUrl", nil)
	}

	conn, err := amqp.Dial(brokerUrl)
	if err != nil {
		return plugin.NewPluginError("connection_failed", fmt.Sprintf("failed to connect to RabbitMQ: %s", brokerUrl), err)
	}
	p.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		return plugin.NewPluginError("connection_failed", "failed to create channel", err)
	}

	prefetchCount := 10
	if val, ok := config["prefetchCount"].(float64); ok {
		prefetchCount = int(val)
	}
	if err := ch.Qos(prefetchCount, 0, false); err != nil {
		return plugin.NewPluginError("internal_error", "failed to set QoS", err)
	}

	queue, ok := config["queue"].(string)
	if !ok || queue == "" {
		return plugin.NewPluginError("config_error", "missing or invalid queue", nil)
	}

	consumerTag, ok := config["consumerTag"].(string)
	if !ok || consumerTag == "" {
		return plugin.NewPluginError("config_error", "missing or invalid consumerTag", nil)
	}

	autoAck := false
	if val, ok := config["autoAck"].(bool); ok {
		autoAck = val
	}

	msgs, err := ch.Consume(
		queue,
		consumerTag,
		autoAck,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return plugin.NewPluginError("internal_error", fmt.Sprintf("failed to register consumer for queue %s", queue), err)
	}

	p.channels[triggerID] = ch

	go func() {
		for msg := range msgs {
			p.handleMessage(ctx, triggerID, msg, autoAck)
		}
	}()

	fmt.Printf("RabbitMQ trigger %s started for queue %s\n", triggerID, queue)

	return nil
}

func (p *RabbitMQTriggerPlugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for triggerID, ch := range p.channels {
		if err := ch.Close(); err != nil {
			fmt.Printf("Failed to close channel for trigger %s: %v\n", triggerID, err)
		} else {
			fmt.Printf("RabbitMQ channel for trigger %s closed\n", triggerID)
		}
	}

	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			fmt.Printf("Failed to close RabbitMQ connection: %v\n", err)
		}
	}

	p.channels = nil
	p.configs = nil
	p.handlers = nil

	return nil
}

func (p *RabbitMQTriggerPlugin) handleMessage(
	ctx context.Context,
	triggerID string,
	msg amqp.Delivery,
	autoAck bool,
) {
	p.mu.RLock()
	handler, handlerExists := p.handlers[triggerID]
	p.mu.RUnlock()

	if !handlerExists {
		fmt.Printf("Handler not found for trigger %s\n", triggerID)
		if !autoAck {
			msg.Nack(false, false)
		}
		return
	}

	var body interface{}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		body = string(msg.Body)
	}

	headers := make(map[string]interface{})
	for k, v := range msg.Headers {
		headers[k] = v
	}

	input := map[string]interface{}{
		"messageId":   msg.MessageId,
		"body":        body,
		"headers":     headers,
		"timestamp":   time.Now().Format(time.RFC3339),
		"deliveryTag": msg.DeliveryTag,
	}

	result, err := handler(ctx, input)

	if !autoAck {
		if err != nil || !result.Success {
			if err := msg.Nack(false, true); err != nil {
				fmt.Printf("Failed to nack message for trigger %s: %v\n", triggerID, err)
			}
		} else {
			if err := msg.Ack(false); err != nil {
				fmt.Printf("Failed to ack message for trigger %s: %v\n", triggerID, err)
			}
		}
	}

	if err != nil {
		fmt.Printf("RabbitMQ trigger %s handler error: %v\n", triggerID, err)
	} else if !result.Success {
		fmt.Printf("RabbitMQ trigger %s handler failed: %s\n", triggerID, result.Error)
	} else {
		fmt.Printf("RabbitMQ trigger %s handled message successfully\n", triggerID)
	}
}
