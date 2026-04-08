<template>
  <div class="run-detail" v-if="run">
    <div class="header">
      <router-link to="/runs" class="back-link">
        <el-icon><ArrowLeft /></el-icon>
        返回列表
      </router-link>
      <h1>运行详情: {{ run.id }}</h1>
      <div class="actions">
        <el-button @click="handleRetry" :disabled="run.status !== 'failed'">
          重试
        </el-button>
      </div>
    </div>

    <!-- 基本信息 -->
    <el-card class="info-card">
      <template #header>
        <span>基本信息</span>
      </template>
      <el-descriptions :column="4" border>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(run.status)">
            {{ statusText(run.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="触发器">
          {{ run.triggerName }}
        </el-descriptions-item>
        <el-descriptions-item label="流程">
          {{ run.flowName }}
        </el-descriptions-item>
        <el-descriptions-item label="耗时">
          {{ formatDuration(run.durationMs) }}
        </el-descriptions-item>
        <el-descriptions-item label="开始时间">
          {{ formatDate(run.startedAt) }}
        </el-descriptions-item>
        <el-descriptions-item label="结束时间" v-if="run.finishedAt">
          {{ formatDate(run.finishedAt) }}
        </el-descriptions-item>
        <el-descriptions-item label="版本">
          v{{ run.flowVersion }}
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- 输入输出 -->
    <div class="io-section">
      <el-card class="io-card">
        <template #header>
          <span>输入数据</span>
        </template>
        <pre>{{ JSON.stringify(run.inputData, null, 2) }}</pre>
      </el-card>

      <el-card class="io-card">
        <template #header>
          <span>输出数据</span>
        </template>
        <pre>{{ JSON.stringify(run.outputData, null, 2) }}</pre>
      </el-card>
    </div>

    <!-- 错误信息 -->
    <el-card v-if="run.errorMessage" class="error-card">
      <template #header>
        <span>错误信息</span>
      </template>
      <el-alert :title="run.errorMessage" type="error" :closable="false" />
    </el-card>

    <!-- 执行时间线 -->
    <ExecutionTimeline
      :timeline="timeline"
      :node-logs="nodeLogs"
      class="timeline-section"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { runAPI } from '@/api/runs'
import ExecutionTimeline from '@/components/run/ExecutionTimeline.vue'
import type { RunDetail, ExecutionTimeline as Timeline } from '@/types/run'
import { formatDate, formatDuration, statusTagType, statusText } from '@/utils/request'

const route = useRoute()
const router = useRouter()

const run = ref<RunDetail>()
const timeline = ref<Timeline>()
const nodeLogs = ref<any[]>([])

async function loadRunDetail() {
  const id = route.params.id as string
  try {
    run.value = await runAPI.get(id)
    timeline.value = await runAPI.getTimeline(id)
    nodeLogs.value = run.value.nodeLogs
  } catch (error) {
    ElMessage.error('加载运行详情失败')
  }
}

async function handleRetry() {
  if (!run.value) return
  
  try {
    await runAPI.retry(run.value.id)
    ElMessage.success('已提交重试')
    loadRunDetail()
  } catch (error) {
    ElMessage.error('重试失败')
  }
}

onMounted(() => {
  loadRunDetail()
})
</script>

<style scoped>
.run-detail {
  padding: 20px;
}

.header {
  display: flex;
  align-items: center;
  margin-bottom: 30px;
}

.back-link {
  display: flex;
  align-items: center;
  color: #409eff;
  text-decoration: none;
  margin-right: 20px;
}

.actions {
  margin-left: auto;
}

.info-card {
  margin-bottom: 30px;
}

.io-section {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 30px;
}

.io-card {
  height: 400px;
  overflow: auto;
}

.error-card {
  margin-bottom: 30px;
}

.timeline-section {
  margin-bottom: 30px;
}

pre {
  background: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
  overflow-x: auto;
  margin: 0;
}
</style>
