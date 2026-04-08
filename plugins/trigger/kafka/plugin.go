package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"mengri-flow/internal/infra/plugin"
)

func init() {
	registry := plugin.GlobalRegistry()
	registry.RegisterTrigger(&KafkaTriggerPlugin{})
}

// 确保接口实现
var _ plugin.TriggerPlugin = (*KafkaTriggerPlugin)(nil)

type KafkaTriggerPlugin struct {
	mu       sync.RWMutex
	readers  map[string]*kafka.Reader
	configs  map[string]map[string]interface{}
	handlers map[string]plugin.TriggerHandler
}

func (p *KafkaTriggerPlugin) PluginMeta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "kafka",
		Type:        plugin.PluginTypeTrigger,
		Version:     "1.0.0",
		Description: "Kafka消息触发器插件",
		Author:      "Platform Team",
	}
}

func (p *KafkaTriggerPlugin) ConfigSchema() plugin.JSONSchema {
	return plugin.JSONSchema{
		"type": "object",
		"required": []string{"brokers", "topic", "groupId"},
		"properties": map[string]interface{}{
			"brokers": map[string]interface{}{
				"type":        "array",
				"title":       "Broker列表",
				"description": "Kafka broker地址列表",
				"items": map[string]interface{}{
					"type": "string",
					"format": "uri",
				},
				"minItems": 1,
			},
			"topic": map[string]interface{}{
				"type":        "string",
				"title":       "Topic",
				"description": "监听的Kafka topic",
				"placeholder": "my-topic",
			},
			"groupId": map[string]interface{}{
				"type":        "string",
				"title":       "消费者组ID",
				"description": "Kafka消费者组ID",
				"placeholder": "my-consumer-group",
			},
			"autoCommit": map[string]interface{}{
				"type":        "boolean",
				"title":       "自动提交偏移量",
				"description": "是否自动提交消费偏移量",
				"default":     true,
			},
			"partition": map[string]interface{}{
				"type":        "integer",
				"title":       "分区",
				"description": "指定分区（不指定则消费所有分区）",
				"minimum":     0,
			},
			"minBytes": map[string]interface{}{
				"type":        "integer",
				"title":       "最小字节数",
				"description": "每次拉取的最小字节数",
				"default":     10240,
				"minimum":     1,
			},
			"maxBytes": map[string]interface{}{
				"type":        "integer",
				"title":       "最大字节数",
				"description": "每次拉取的最大字节数",
				"default":     10485760,
				"minimum":     1,
			},
			"maxWait": map[string]interface{}{
				"type":        "integer",
				"title":       "最大等待时间",
				"description": "等待消息的最大时间（毫秒）",
				"default":     1000,
				"minimum":     1,
			},
		},
	}
}

func (p *KafkaTriggerPlugin) InputSchema() plugin.JSONSchema {
	return plugin.JSONSchema{
		"type": "object",
		"properties": map[string]interface{}{
			"key": map[string]interface{}{
				"type":        "string",
				"description": "消息Key",
			},
			"value": map[string]interface{}{
				"type":        "object",
				"description": "消息Value（JSON解析后）",
			},
			"topic": map[string]interface{}{
				"type":        "string",
				"description": "消息Topic",
			},
			"partition": map[string]interface{}{
				"type":        "integer",
				"description": "消息分区",
			},
			"offset": map[string]interface{}{
				"type":        "integer",
				"description": "消息偏移量",
			},
			"timestamp": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "消息时间戳",
			},
		},
	}
}

func (p *KafkaTriggerPlugin) OutputSchema() plugin.JSONSchema {
	return nil
}

func (p *KafkaTriggerPlugin) Start(
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

	if p.readers == nil {
		p.readers = make(map[string]*kafka.Reader)
		p.configs = make(map[string]map[string]interface{})
		p.handlers = make(map[string]plugin.TriggerHandler)
	}

	p.configs[triggerID] = config
	p.handlers[triggerID] = handler

	brokersInterface, ok := config["brokers"].([]interface{})
	if !ok || len(brokersInterface) == 0 {
		return plugin.NewPluginError("config_error", "missing or invalid brokers", nil)
	}

	brokers := make([]string, len(brokersInterface))
	for i, b := range brokersInterface {
		if broker, ok := b.(string); ok {
			brokers[i] = broker
		} else {
			return plugin.NewPluginError("config_error", "invalid broker format", nil)
		}
	}

	topic, ok := config["topic"].(string)
	if !ok || topic == "" {
		return plugin.NewPluginError("config_error", "missing or invalid topic", nil)
	}

	groupId, ok := config["groupId"].(string)
	if !ok || groupId == "" {
		return plugin.NewPluginError("config_error", "missing or invalid groupId", nil)
	}

	readerConfig := kafka.ReaderConfig{
		Brokers:   brokers,
		Topic:     topic,
		GroupID:   groupId,
		MinBytes:  10240,
		MaxBytes:  10485760,
		MaxWait:   1 * time.Second,
	}

	if val, ok := config["minBytes"].(float64); ok {
		readerConfig.MinBytes = int(val)
	}
	if val, ok := config["maxBytes"].(float64); ok {
		readerConfig.MaxBytes = int(val)
	}
	if val, ok := config["maxWait"].(float64); ok {
		readerConfig.MaxWait = time.Duration(val) * time.Millisecond
	}

	if partition, ok := config["partition"].(float64); ok {
		readerConfig.Partition = int(partition)
		readerConfig.GroupID = ""
	}

	autoCommit := true
	if val, ok := config["autoCommit"].(bool); ok {
		autoCommit = val
	}
	readerConfig.CommitInterval = 0
	if autoCommit {
		readerConfig.CommitInterval = 1 * time.Second
	}

	reader := kafka.NewReader(readerConfig)

	p.readers[triggerID] = reader

	go func() {
		for {
			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				if err != context.Canceled {
					fmt.Printf("Kafka trigger %s fetch message error: %v\n", triggerID, err)
				}
				return
			}

			p.handleKafkaMessage(ctx, triggerID, msg, autoCommit)
		}
	}()

	fmt.Printf("Kafka trigger %s started for topic %s, group %s\n", triggerID, topic, groupId)

	return nil
}

func (p *KafkaTriggerPlugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for triggerID, reader := range p.readers {
		if err := reader.Close(); err != nil {
			fmt.Printf("Failed to close Kafka reader for trigger %s: %v\n", triggerID, err)
		} else {
			fmt.Printf("Kafka reader for trigger %s closed\n", triggerID)
		}
	}

	p.readers = nil
	p.configs = nil
	p.handlers = nil

	return nil
}

func (p *KafkaTriggerPlugin) handleKafkaMessage(
	ctx context.Context,
	triggerID string,
	msg kafka.Message,
	autoCommit bool,
) {
	p.mu.RLock()
	handler, handlerExists := p.handlers[triggerID]
	p.mu.RUnlock()

	if !handlerExists {
		fmt.Printf("Handler not found for trigger %s\n", triggerID)
		return
	}

	var value interface{}
	if err := json.Unmarshal(msg.Value, &value); err != nil {
		value = string(msg.Value)
	}

	input := map[string]interface{}{
		"key":       string(msg.Key),
		"value":     value,
		"topic":     msg.Topic,
		"partition": msg.Partition,
		"offset":    msg.Offset,
		"timestamp": msg.Time.Format(time.RFC3339),
	}

	result, err := handler(ctx, input)

	if !autoCommit {
		if err != nil || !result.Success {
			fmt.Printf("Kafka trigger %s handler failed, not committing offset\n", triggerID)
		} else {
			reader := p.readers[triggerID]
			if reader != nil {
				if err := reader.CommitMessages(ctx, msg); err != nil {
					fmt.Printf("Failed to commit message for trigger %s: %v\n", triggerID, err)
				}
			}
		}
	}

	if err != nil {
		fmt.Printf("Kafka trigger %s handler error: %v\n", triggerID, err)
	} else if !result.Success {
		fmt.Printf("Kafka trigger %s handler failed: %s\n", triggerID, result.Error)
	} else {
		fmt.Printf("Kafka trigger %s handled message successfully (offset: %d)\n", triggerID, msg.Offset)
	}
}
