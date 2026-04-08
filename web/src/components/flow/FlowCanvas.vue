<template>
  <div class="flow-canvas" ref="canvasRef">
    <!-- 工具面板 -->
    <NodePanel
      :tools="tools"
      @drag-start="handleDragStart"
    />

    <!-- 画布区域 -->
    <div class="canvas-area" @drop="handleDrop" @dragover.prevent>
      <VueFlow
        v-model="flowData"
        :default-zoom="1"
        :min-zoom="0.5"
        :max-zoom="2"
        :snap-to-grid="true"
        :snap-grid="[20, 20]"
        @node-click="handleNodeClick"
        @connect="handleConnect"
      >
        <template #node-start="{ node }">
          <StartNode :node="node" />
        </template>

        <template #node-end="{ node }">
          <EndNode :node="node" />
        </template>

        <template #node-tool="{ node }">
          <ToolNode
            :node="node"
            :status="nodeStatus[node.id]"
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

        <!-- 迷你地图 -->
        <MiniMap />
      </VueFlow>
    </div>

    <!-- 配置抽屉 -->
    <ConfigDrawer
      v-model="drawerVisible"
      :node="selectedNode"
      :flow-input-schema="flow.inputSchema"
      @save="handleNodeConfigSave"
    />

    <!-- 测试面板 -->
    <TestPanel
      v-model="testPanelVisible"
      :flow="flow"
      @run="handleRunTest"
      @close="testPanelVisible = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { VueFlow, useVueFlow, Controls, ControlButton, MiniMap } from '@vue-flow/core'
import { useFlowCanvas } from '@/composables/useFlowCanvas'
import NodePanel from './NodePanel.vue'
import ConfigDrawer from './ConfigDrawer.vue'
import TestPanel from './TestPanel.vue'
import { toolAPI } from '@/api/tools'
import { flowAPI } from '@/api/flows'

const props = defineProps<{
  flowId: string
}>()

const {
  nodes,
  edges,
  addNodes,
  addEdges,
  removeNodes,
  updateNode,
  fitView,
} = useVueFlow()

const {
  flowData,
  selectedNode,
  drawerVisible,
  testPanelVisible,
  nodeStatus,
  loadFlow,
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
  tools.value = await toolAPI.list({
    workspaceId: workspaceStore.currentWorkspace,
    status: 'published',
  })
})

async function handleSave() {
  await flowAPI.update(props.flowId, {
    canvasData: flowData.value,
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
    nodeStatus.value[log.nodeId] = log.status
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
