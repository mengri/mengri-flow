<template>
  <div class="flow-canvas">
    <!-- 工具面板 -->
    <NodePanel
      :tools="tools"
      @drag-start="handleDragStart"
    />

    <!-- 画布区域 -->
    <div class="canvas-area" @drop="handleDrop" @dragover.prevent>
      <VueFlow
        v-model:nodes="nodes"
        v-model:edges="edges"
        :default-zoom="1"
        :min-zoom="0.5"
        :max-zoom="2"
        :snap-to-grid="true"
        :snap-grid="[20, 20]"
        @node-click="handleNodeClick"
        @connect="handleConnect"
      >
        <template #node-start="nodeProps">
          <StartNode :node="(nodeProps as any).node" />
        </template>

        <template #node-end="nodeProps">
          <EndNode :node="(nodeProps as any).node" />
        </template>

        <template #node-tool="nodeProps">
          <ToolNode
            :node="(nodeProps as any).node"
            :status="nodeStatus[(nodeProps as any).node.id]"
            @remove="removeNode"
          />
        </template>

        <!-- 控制按钮 -->
        <Controls>
          <ControlButton title="保存" @click="handleSave">
            <el-icon><Check /></el-icon>
          </ControlButton>
          <ControlButton title="测试运行" @click="handleTestRun">
            <el-icon><VideoPlay /></el-icon>
          </ControlButton>
          <ControlButton title="发布" @click="handlePublish">
            <el-icon><Upload /></el-icon>
          </ControlButton>
          <ControlButton title="适配视图" @click="fitView">
            <el-icon><FullScreen /></el-icon>
          </ControlButton>
        </Controls>
      </VueFlow>
    </div>

    <!-- 配置抽屉 -->
    <ConfigDrawer
      v-model="drawerVisible"
      :node="selectedNode"
      :flow-input-schema="flowData?.inputSchema || {}"
      @save="handleNodeConfigSave"
    />

    <!-- 测试面板 -->
    <TestPanel
      v-if="flowData"
      v-model="testPanelVisible"
      :flow="flowData"
      @run="handleRunTest"
      @close="testPanelVisible = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, type Ref } from 'vue'
import { VueFlow, useVueFlow, type GraphNode, type GraphEdge } from '@vue-flow/core'
import { Controls, ControlButton } from '@vue-flow/controls'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useWorkspaceStore } from '@/stores/workspace'
import { useFlowCanvas } from '@/composables/useFlowCanvas'
import NodePanel from './NodePanel.vue'
import ConfigDrawer from './ConfigDrawer.vue'
import TestPanel from './TestPanel.vue'
import { toolAPI } from '@/api/tools'
import { flowAPI } from '@/api/flows'
import type { Tool } from '@/types/tool'

const props = defineProps<{
  flowId: string
}>()

const {
  nodes,
  edges,
  fitView,
} = useVueFlow() as { nodes: Ref<GraphNode[]>, edges: Ref<GraphEdge[]>, fitView: () => void }

const {
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
} = useFlowCanvas(props.flowId)

const tools = ref<Tool[]>([])

onMounted(async () => {
  // 加载工具列表
  const workspaceStore = useWorkspaceStore()
  const workspaceId = workspaceStore.currentWorkspaceIdOrThrow
  tools.value = await toolAPI.list({
    workspaceId,
    status: 'published',
  })
})

async function handleSave() {
  if (!flowData.value?.canvasData) return
  await flowAPI.update(props.flowId, {
    canvasData: {
      nodes: nodes.value.map(node => ({
        id: node.id,
        type: node.type as 'start' | 'end' | 'tool' | 'condition',
        position: node.position,
        data: node.data,
      })),
      edges: edges.value.map(edge => ({
        id: edge.id,
        source: edge.source,
        target: edge.target,
        type: edge.type || 'default',
      })),
    },
  })
  ElMessage.success('保存成功')
}

function handleTestRun() {
  testPanelVisible.value = true
}

async function handlePublish() {
  await ElMessageBox.confirm('确定要发布该流程吗？', '确认')
  await flowAPI.publish(props.flowId)
  ElMessage.success('发布成功')
}

async function handleRunTest(input: Record<string, any>) {
  const result = await flowAPI.test({
    flowId: props.flowId,
    input,
  })
  
  // 更新节点状态
  result.nodeLogs.forEach((log: any) => {
    ;(nodeStatus as any)[log.nodeId] = log.status
  })
  
  return result
}
</script>

<style scoped>
.flow-canvas {
  display: flex;
  height: 100vh;
  position: relative;
}

.canvas-area {
  flex: 1;
  height: 100%;
}
</style>
