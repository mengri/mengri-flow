import { ref, reactive } from 'vue'
import { useVueFlow, type GraphNode, type GraphEdge } from '@vue-flow/core'
import type { Flow } from '@/types/flow'
import type { Tool } from '@/types/tool'
import { flowAPI } from '@/api/flows'

export function useFlowCanvas(flowId: string) {
  const flowData = ref<Flow>()
  const selectedNode = ref<GraphNode | null>(null)
  const drawerVisible = ref(false)
  const testPanelVisible = ref(false)
  const nodeStatus = reactive<Record<string, string>>({})

  const { nodes, edges, addNodes, addEdges, removeNodes, updateNode } = useVueFlow()

  async function loadFlow() {
    const result = await flowAPI.get(flowId)
    flowData.value = result
    
    if (!result.canvasData) return
    
    // 初始化节点和连线 - 使用类型断言避免复杂类型问题
    nodes.value = result.canvasData.nodes.map((node: any) => ({
      id: node.id,
      type: node.type,
      position: node.position,
      data: node.data,
    })) as GraphNode[]
    
    edges.value = result.canvasData.edges.map((edge: any) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      type: edge.type,
    })) as GraphEdge[]
  }

  // 加载流程
  loadFlow()

  function handleNodeClick(event: any) {
    const node = event.node as GraphNode
    if (node.type === 'tool') {
      selectedNode.value = node
      drawerVisible.value = true
    }
  }

  function handleConnect(params: any) {
    addEdges([params])
  }

  function handleDrop(event: DragEvent) {
    const toolData = JSON.parse(event.dataTransfer?.getData('tool') || '{}')
    
    if (!toolData.id) return
    
    const rect = (event.target as HTMLElement).getBoundingClientRect()
    const position = {
      x: event.clientX - rect.left,
      y: event.clientY - rect.top,
    }
    
    const newNode = {
      id: `node_${Date.now()}`,
      type: 'tool',
      position,
      data: {
        toolId: toolData.id,
        toolName: toolData.name,
        toolVersion: toolData.currentVersion,
        inputMapping: [],
        config: {
          timeout: 30000,
          retry: 0,
          condition: '',
        },
      },
    }
    
    addNodes([newNode as GraphNode])
  }

  function handleDragStart(event: DragEvent, tool: Tool) {
    event.dataTransfer?.setData('tool', JSON.stringify(tool))
  }

  function handleNodeConfigSave(nodeId: string, config: any) {
    updateNode(nodeId, (node) => {
      node.data = { ...node.data, ...config }
      return node
    })
    drawerVisible.value = false
    selectedNode.value = null
  }

  function removeNode(nodeId: string) {
    removeNodes([nodeId])
  }

  return {
    flowData,
    selectedNode,
    drawerVisible,
    testPanelVisible,
    nodeStatus,
    handleNodeClick,
    handleConnect,
    handleDrop,
    handleDragStart,
    removeNode,
    handleNodeConfigSave,
  }
}
