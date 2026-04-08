<template>
  <el-card class="execution-timeline">
    <template #header>
      <span>执行时间线</span>
    </template>

    <div class="timeline-container">
      <el-timeline>
        <el-timeline-item
          v-for="item in timeline"
          :key="item.id"
          :timestamp="formatDate(item.timestamp)"
          :type="itemType(item.event)"
        >
          <div class="timeline-item">
            <div class="event">{{ eventText(item.event) }}</div>
            <div v-if="item.nodeId" class="node">
              节点: {{ getNodeName(item.nodeId) }}
            </div>
            <div v-if="item.duration" class="duration">
              耗时: {{ formatDuration(item.duration) }}
            </div>
          </div>
        </el-timeline-item>
      </el-timeline>
    </div>

    <!-- 节点日志详情 -->
    <div class="node-logs">
      <h4>节点执行日志</h4>
      <el-table :data="nodeLogs" border>
        <el-table-column prop="nodeId" label="节点" width="120" />
        <el-table-column prop="toolName" label="工具" width="150" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">
              {{ statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="durationMs" label="耗时" width="100">
          <template #default="{ row }">
            {{ formatDuration(row.durationMs) }}
          </template>
        </el-table-column>
        <el-table-column label="输入/输出" width="120">
          <template #default="{ row }">
            <el-button size="small" @click="showNodeDetails(row)">
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ExecutionTimeline, NodeLog } from '@/types/run'
import { formatDate, formatDuration, statusTagType, statusText } from '@/utils/request'

const props = defineProps<{
  timeline: ExecutionTimeline
  nodeLogs: NodeLog[]
}>()

const timeline = computed(() => {
  return props.timeline.timeline.sort(
    (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
  )
})

function itemType(event: string) {
  if (event.includes('failed')) return 'danger'
  if (event.includes('success')) return 'success'
  return ''
}

function eventText(event: string) {
  const map: Record<string, string> = {
    flow_started: '流程开始执行',
    flow_success: '流程执行成功',
    flow_failed: '流程执行失败',
    node_started: '节点开始执行',
    node_success: '节点执行成功',
    node_failed: '节点执行失败',
  }
  return map[event] || event
}

function getNodeName(nodeId: string) {
  const nodeLog = props.nodeLogs.find(log => log.nodeId === nodeId)
  return nodeLog?.toolName || nodeId
}

function showNodeDetails(row: NodeLog) {
  // 显示节点详情弹窗
  console.log('Node details:', row)
}
</script>

<style scoped>
.execution-timeline {
  margin-bottom: 30px;
}

.timeline-container {
  padding: 20px 0;
}

.timeline-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.event {
  font-weight: 500;
  color: #303133;
}

.node {
  font-size: 14px;
  color: #606266;
}

.duration {
  font-size: 12px;
  color: #909399;
}

.node-logs {
  margin-top: 30px;
  border-top: 1px solid #e4e7ed;
  padding-top: 20px;
}

.node-logs h4 {
  margin-bottom: 15px;
  color: #303133;
}
</style>
