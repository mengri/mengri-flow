package executor

import (
	"context"
	"fmt"
	"log"
	"sync"

	"mengri-flow/internal/domain/entity"
	"mengri-flow/internal/infra/plugin"
)

// TriggerManager 触发器管理器
type TriggerManager struct {
	flowEngine *FlowEngine
	triggers   map[string]plugin.TriggerPlugin
	handlers   map[string]plugin.TriggerHandler
	mu         sync.RWMutex
	nodeID     string
}

// NewTriggerManager 创建新的触发器管理器
func NewTriggerManager(flowEngine *FlowEngine, nodeID string) *TriggerManager {
	return &TriggerManager{
		flowEngine: flowEngine,
		triggers:   make(map[string]plugin.TriggerPlugin),
		handlers:   make(map[string]plugin.TriggerHandler),
		nodeID:     nodeID,
	}
}

// AddOrUpdateTrigger 添加或更新触发器
func (m *TriggerManager) AddOrUpdateTrigger(trigger *entity.Trigger) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	triggerID := trigger.ID.String()
	log.Printf("Adding/Updating trigger %s of type %s", triggerID, trigger.Type)

	// 1. 如果已存在，先停止旧的
	if oldPlugin, exists := m.triggers[triggerID]; exists {
		log.Printf("Stopping old trigger %s", triggerID)
		oldPlugin.Stop()
		delete(m.triggers, triggerID)
		delete(m.handlers, triggerID)
	}

	// 2. 获取触发器插件（这里简化处理，实际需要插件注册表）
	// 暂时直接创建插件实例
	plugin, err := m.createTriggerPlugin(trigger.Type)
	if err != nil {
		return fmt.Errorf("failed to create trigger plugin: %w", err)
	}

	// 3. 创建handler
	handler := m.createTriggerHandler(trigger)

	// 4. 启动触发器
	config := make(map[string]interface{})
	for k, v := range trigger.Config {
		config[k] = v
	}
	config["triggerId"] = triggerID
	config["nodeID"] = m.nodeID

	ctx := context.Background()
	if err := plugin.Start(ctx, config, handler); err != nil {
		return fmt.Errorf("failed to start trigger: %w", err)
	}

	// 5. 保存
	m.triggers[triggerID] = plugin
	m.handlers[triggerID] = handler

	log.Printf("Trigger %s started successfully", triggerID)

	return nil
}

// RemoveTrigger 移除触发器
func (m *TriggerManager) RemoveTrigger(triggerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("Removing trigger %s", triggerID)

	if plugin, exists := m.triggers[triggerID]; exists {
		plugin.Stop()
		delete(m.triggers, triggerID)
		delete(m.handlers, triggerID)
		log.Printf("Trigger %s removed", triggerID)
	}
}

// StopAll 停止所有触发器
func (m *TriggerManager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("Stopping all triggers")

	for triggerID, plugin := range m.triggers {
		log.Printf("Stopping trigger %s", triggerID)
		plugin.Stop()
	}

	m.triggers = make(map[string]plugin.TriggerPlugin)
	m.handlers = make(map[string]plugin.TriggerHandler)

	log.Printf("All triggers stopped")
}

// createTriggerPlugin 创建触发器插件（简化版本）
func (m *TriggerManager) createTriggerPlugin(triggerType entity.TriggerType) (plugin.TriggerPlugin, error) {
	switch triggerType {
	case entity.TriggerTypeRESTful:
		// 返回 RESTful 触发器插件
		return nil, fmt.Errorf("RESTful trigger plugin not implemented yet")
	case entity.TriggerTypeTimer:
		// 返回 Timer 触发器插件
		return nil, fmt.Errorf("Timer trigger plugin not implemented yet")
	case entity.TriggerTypeMQ:
		// 返回 MQ 触发器插件
		return nil, fmt.Errorf("MQ trigger plugin not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported trigger type: %s", triggerType)
	}
}

// createTriggerHandler 创建触发器处理器
func (m *TriggerManager) createTriggerHandler(trigger *entity.Trigger) plugin.TriggerHandler {
	return func(ctx context.Context, input map[string]any) (*plugin.TriggerResult, error) {
		triggerID := trigger.ID.String()
		log.Printf("Trigger %s executed", triggerID)

		// 1. 输入映射
		flowInput := m.applyInputMapping(input, trigger.InputMapping)
		log.Printf("Mapped input for trigger %s: %v", triggerID, flowInput)

		// 2. 查询流程
		flow, err := m.flowEngine.GetFlow(trigger.FlowID.String())
		if err != nil {
			log.Printf("Failed to get flow %s: %v", trigger.FlowID, err)
			return m.handleError(err, trigger.ErrorHandling)
		}

		log.Printf("Executing flow %s for trigger %s", flow.ID, triggerID)

		// 3. 执行流程
		ctxContext := context.Background()
		if ctx != nil {
			if c, ok := ctx.(context.Context); ok {
				ctxContext = c
			}
		}
		output, err := m.flowEngine.ExecuteFlow(ctxContext, flow, flowInput)
		if err != nil {
			log.Printf("Flow execution failed for trigger %s: %v", triggerID, err)
			return m.handleError(err, trigger.ErrorHandling)
		}

		log.Printf("Flow executed successfully for trigger %s", triggerID)

		// 4. 输出映射
		response := m.applyOutputMapping(output, trigger.OutputMapping)

		return &plugin.TriggerResult{
			Success: true,
			Data:    response,
		}, nil
	}
}

// applyInputMapping 应用输入映射
func (m *TriggerManager) applyInputMapping(input map[string]interface{}, mapping map[string]interface{}) map[string]interface{} {
	if mapping == nil {
		return input
	}

	result := make(map[string]interface{})

	for key, value := range mapping {
		if strValue, ok := value.(string); ok {
			// 简单实现：支持直接字段映射
			if inputValue, exists := input[strValue]; exists {
				result[key] = inputValue
			}
		} else {
			result[key] = value
		}
	}

	return result
}

// applyOutputMapping 应用输出映射
func (m *TriggerManager) applyOutputMapping(output map[string]interface{}, mapping map[string]interface{}) map[string]interface{} {
	if mapping == nil {
		return output
	}

	result := make(map[string]interface{})

	for key, value := range mapping {
		if strValue, ok := value.(string); ok {
			if outputValue, exists := output[strValue]; exists {
				result[key] = outputValue
			}
		} else {
			result[key] = value
		}
	}

	return result
}

// handleError 处理错误
func (m *TriggerManager) handleError(err error, errorHandling map[string]interface{}) (*plugin.TriggerResult, error) {
	strategy := "fail"
	if errorHandling != nil {
		if s, ok := errorHandling["strategy"].(string); ok {
			strategy = s
		}
	}

	switch strategy {
	case "ignore":
		// 忽略错误，返回空结果
		return &plugin.TriggerResult{
			Success: true,
			Data:    map[string]interface{}{},
		}, nil
	case "retry":
		// 重试逻辑（简化版）
		log.Printf("Retry not implemented yet, failing instead")
		fallthrough
	default:
		// 默认失败
		return &plugin.TriggerResult{
			Success: false,
			Error:   err.Error(),
			Data:    map[string]interface{}{},
		}, nil
	}
}
