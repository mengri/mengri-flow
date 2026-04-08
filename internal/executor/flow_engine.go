package executor

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"mengri-flow/internal/domain/entity"
)

// FlowEngine 流程执行引擎
type FlowEngine struct {
	nodeID        string
	flowCache     map[string]*entity.Flow
	toolCache     map[string]*entity.Tool
	resourceCache map[string]*entity.Resource
	mu            sync.RWMutex
}

// NewFlowEngine 创建新的流程执行引擎
func NewFlowEngine(nodeID string) *FlowEngine {
	return &FlowEngine{
		nodeID:        nodeID,
		flowCache:     make(map[string]*entity.Flow),
		toolCache:     make(map[string]*entity.Tool),
		resourceCache: make(map[string]*entity.Resource),
	}
}

// GetFlow 获取流程
func (e *FlowEngine) GetFlow(flowID string) (*entity.Flow, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	flow, exists := e.flowCache[flowID]
	if !exists {
		return nil, fmt.Errorf("flow not found: %s", flowID)
	}

	return flow, nil
}

// CacheFlow 缓存流程
func (e *FlowEngine) CacheFlow(flow *entity.Flow) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.flowCache[flow.ID.String()] = flow
	log.Printf("Flow %s cached", flow.ID)
}

// ExecuteFlow 执行流程
func (e *FlowEngine) ExecuteFlow(ctx context.Context, flow *entity.Flow, input map[string]interface{}) (map[string]interface{}, error) {
	log.Printf("Executing flow %s", flow.ID)

	// 1. 构建执行计划
	plan, err := e.buildExecutionPlan(flow.CanvasData)
	if err != nil {
		return nil, fmt.Errorf("failed to build execution plan: %w", err)
	}

	log.Printf("Execution plan built with %d nodes", len(plan))

	// 2. 执行节点
	contextData := make(map[string]interface{})
	contextData["start"] = input

	for _, node := range plan {
		log.Printf("Executing node %s of type %s", node.ID, node.Type)

		output, err := e.executeNode(ctx, node, contextData)
		if err != nil {
			return nil, fmt.Errorf("failed to execute node %s: %w", node.ID, err)
		}

		contextData[node.ID] = output
		log.Printf("Node %s executed successfully", node.ID)
	}

	// 3. 返回结果
	endOutput, exists := contextData["end"]
	if !exists {
		return nil, fmt.Errorf("no end node output found")
	}

	result, ok := endOutput.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid end node output type")
	}

	log.Printf("Flow %s executed successfully", flow.ID)

	return result, nil
}

// Node 流程节点
type Node struct {
	ID       string
	Type     string
	ToolID   string
	Config   map[string]interface{}
	Input    map[string]interface{}
	Position map[string]interface{}
}

// buildExecutionPlan 构建执行计划
func (e *FlowEngine) buildExecutionPlan(canvasData map[string]interface{}) ([]*Node, error) {
	// 简化实现：假设画布数据包含节点和连接信息
	// 实际需要从 CanvasData 解析节点和边，然后拓扑排序

	nodes := make([]*Node, 0)

	// 示例：假设有一个开始节点、一个工具节点和一个结束节点
	nodes = append(nodes, &Node{
		ID:   "start",
		Type: "start",
	})

	// 从画布数据中提取工具节点
	if nodesData, ok := canvasData["nodes"].([]interface{}); ok {
		for _, nodeData := range nodesData {
			if nodeMap, ok := nodeData.(map[string]interface{}); ok {
				node := &Node{
					ID:   getStringValue(nodeMap, "id"),
					Type: getStringValue(nodeMap, "type"),
				}

				if data, ok := nodeMap["data"].(map[string]interface{}); ok {
					node.ToolID = getStringValue(data, "toolId")
					node.Config = getMapValue(data, "config")
					node.Input = getMapValue(data, "input")
				}

				nodes = append(nodes, node)
			}
		}
	}

	nodes = append(nodes, &Node{
		ID:   "end",
		Type: "end",
	})

	return nodes, nil
}

// executeNode 执行节点
func (e *FlowEngine) executeNode(ctx context.Context, node *Node, contextData map[string]interface{}) (map[string]interface{}, error) {
	switch node.Type {
	case "start":
		// 开始节点：返回输入数据
		if input, exists := contextData["start"]; exists {
			if inputMap, ok := input.(map[string]interface{}); ok {
				return inputMap, nil
			}
		}
		return map[string]interface{}{}, nil

	case "tool":
		return e.executeToolNode(ctx, node, contextData)

	case "end":
		// 结束节点：收集所有上游节点的输出
		result := make(map[string]interface{})
		for k, v := range contextData {
			if k != "start" && k != "end" {
				result[k] = v
			}
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unsupported node type: %s", node.Type)
	}
}

// executeToolNode 执行工具节点
func (e *FlowEngine) executeToolNode(ctx context.Context, node *Node, contextData map[string]interface{}) (map[string]interface{}, error) {
	log.Printf("Executing tool node %s with tool %s", node.ID, node.ToolID)

	// 1. 获取工具
	tool, exists := e.toolCache[node.ToolID]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", node.ToolID)
	}

	// 2. 获取资源
	resource, exists := e.resourceCache[tool.ResourceID.String()]
	if !exists {
		return nil, fmt.Errorf("resource not found: %s", tool.ResourceID)
	}

	log.Printf("Using resource %s of type %s", resource.ID, resource.Type)

	// 3. 参数映射
	input := e.mapInput(node, contextData)

	// 4. 执行工具（简化版：直接返回输入作为输出）
	// 实际需要调用插件的 ExecuteTool 方法
	output := e.executeTool(ctx, tool, resource, input)

	log.Printf("Tool node %s executed successfully", node.ID)

	return output, nil
}

// executeTool 执行工具（简化版）
func (e *FlowEngine) executeTool(ctx context.Context, tool *entity.Tool, resource *entity.Resource, input map[string]interface{}) map[string]interface{} {
	// 简化实现：直接返回输入作为输出
	// 实际应该调用插件的 ExecuteTool 方法
	result := make(map[string]interface{})

	// 合并工具配置和输入
	for k, v := range tool.Config {
		result[k] = v
	}
	for k, v := range input {
		result[k] = v
	}

	return result
}

// mapInput 映射输入参数
func (e *FlowEngine) mapInput(node *Node, contextData map[string]interface{}) map[string]interface{} {
	if node.Input == nil {
		return map[string]interface{}{}
	}

	result := make(map[string]interface{})

	for key, value := range node.Input {
		if strValue, ok := value.(string); ok {
			// 支持引用其他节点的输出
			if strings.HasPrefix(strValue, "${") && strings.HasSuffix(strValue, "}") {
				ref := strValue[2 : len(strValue)-1]
				if refValue, exists := contextData[ref]; exists {
					result[key] = refValue
				}
			} else {
				result[key] = value
			}
		} else {
			result[key] = value
		}
	}

	return result
}

// Helper functions
func getStringValue(m map[string]interface{}, key string) string {
	if v, exists := m[key]; exists {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getMapValue(m map[string]interface{}, key string) map[string]interface{} {
	if v, exists := m[key]; exists {
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}
