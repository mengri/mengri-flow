<template>
  <div class="node-panel">
    <div class="panel-header">
      <h3>工具节点</h3>
    </div>
    
    <div class="panel-content">
      <div
        v-for="tool in tools"
        :key="tool.id"
        class="node-item"
        draggable="true"
        @dragstart="(e) => handleDragStart(e, tool)"
      >
        <div class="node-icon">
          <el-icon><Tools /></el-icon>
        </div>
        <div class="node-info">
          <div class="node-name">{{ tool.name }}</div>
          <div class="node-type">{{ tool.type }}</div>
        </div>
      </div>
    </div>

    <!-- 基础节点 -->
    <div class="panel-header" style="margin-top: 20px">
      <h3>基础节点</h3>
    </div>
    
    <div class="panel-content">
      <div
        v-for="node in basicNodes"
        :key="node.type"
        class="node-item basic-node"
        draggable="true"
        @dragstart="(e) => handleBasicNodeDragStart(e, node)"
      >
        <div class="node-icon">
          <el-icon :size="24">
            <component :is="node.icon" />
          </el-icon>
        </div>
        <div class="node-info">
          <div class="node-name">{{ node.name }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Tool } from '@/types/tool'
import { Tools } from '@element-plus/icons-vue'

defineProps<{
  tools: Tool[]
}>()

const emit = defineEmits<{
  (e: 'drag-start', event: DragEvent, tool: Tool): void
  (e: 'basic-node-drag-start', event: DragEvent, node: any): void
}>()

const basicNodes = [
  {
    type: 'start',
    name: '开始',
    icon: 'CircleCheck',
    color: '#67c23a',
  },
  {
    type: 'end',
    name: '结束',
    icon: 'CircleClose',
    color: '#f56c6c',
  },
  {
    type: 'condition',
    name: '条件判断',
    icon: 'VideoPause',
    color: '#e6a23c',
  },
]

function handleDragStart(event: DragEvent, tool: Tool) {
  emit('drag-start', event, tool)
}

function handleBasicNodeDragStart(event: DragEvent, node: any) {
  emit('basic-node-drag-start', event, node)
}
</script>

<style scoped>
.node-panel {
  width: 250px;
  background: #f5f7fa;
  border-right: 1px solid #dcdfe6;
  padding: 15px;
  overflow-y: auto;
}

.panel-header {
  margin-bottom: 15px;
}

.panel-header h3 {
  margin: 0;
  font-size: 16px;
  color: #303133;
}

.node-item {
  background: white;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 12px;
  margin-bottom: 10px;
  cursor: move;
  transition: all 0.3s;
  display: flex;
  align-items: center;
}

.node-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.2);
}

.node-icon {
  margin-right: 12px;
  color: #409eff;
}

.node-info {
  flex: 1;
}

.node-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.node-type {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

.basic-node .node-icon {
  color: #67c23a;
}
</style>
