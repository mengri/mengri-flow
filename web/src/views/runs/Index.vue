<template>
  <div class="run-list">
    <div class="header">
      <h1>运行记录</h1>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-cards">
      <el-card v-for="stat in stats" :key="stat.title">
        <template #header>
          <span>{{ stat.title }}</span>
        </template>
        <div class="stat-value">{{ stat.value }}</div>
        <div class="stat-change" :class="stat.change > 0 ? 'up' : 'down'">
          {{ stat.change > 0 ? '+' : '' }}{{ stat.change }}%
        </div>
      </el-card>
    </div>

    <!-- 筛选 -->
    <div class="filters">
      <el-form :inline="true" :model="filters">
        <el-form-item>
          <el-input
            v-model="filters.search"
            placeholder="搜索运行ID或流程名称"
            clearable
            style="width: 300px"
          />
        </el-form-item>
        <el-form-item>
          <el-select v-model="filters.status" placeholder="状态" clearable>
            <el-option label="成功" value="success" />
            <el-option label="失败" value="failed" />
            <el-option label="运行中" value="running" />
            <el-option label="超时" value="timeout" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-date-picker
            v-model="filters.dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 图表 -->
    <div class="charts">
      <el-row :gutter="20">
        <el-col :span="16">
          <el-card class="chart-card">
            <template #header>
              <span>运行趋势（7天）</span>
            </template>
            <div class="chart-container">
              <canvas ref="trendChartRef"></canvas>
            </div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card class="chart-card">
            <template #header>
              <span>状态分布</span>
            </template>
            <div class="chart-container chart-container--small">
              <canvas ref="statusChartRef"></canvas>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 表格 -->
    <el-table :data="runs" v-loading="loading" border>
      <el-table-column prop="id" label="运行ID" width="180" fixed="left">
        <template #default="{ row }">
          <router-link :to="runDetailPath(row.id)" class="run-id-link">
            {{ row.id }}
          </router-link>
        </template>
      </el-table-column>
      <el-table-column prop="triggerName" label="触发器" width="150" />
      <el-table-column prop="flowName" label="流程" min-width="200" />
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
      <el-table-column prop="startedAt" label="开始时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.startedAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleView(row.id)">查看</el-button>
          <el-button
            size="small"
            :disabled="row.status !== 'failed'"
            @click="handleRetry(row.id)"
          >
            重试
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Chart, registerables } from 'chart.js'
import { runAPI } from '@/api/runs'
import type { Run, RunStats } from '@/types/run'
import { useWorkspaceStore } from '@/stores/workspace'
import { useWorkspaceRoute } from '@/composables/useWorkspaceRoute'
import { formatDate, formatDuration, statusTagType, statusText } from '@/utils/request'

Chart.register(...registerables)

const router = useRouter()
const workspaceStore = useWorkspaceStore()
const { runDetailPath } = useWorkspaceRoute()

const loading = ref(false)
const runs = ref<Run[]>([])
const stats = ref<any[]>([])
const runStats = ref<RunStats | null>(null)

const trendChartRef = ref<HTMLCanvasElement>()
const statusChartRef = ref<HTMLCanvasElement>()
let trendChart: Chart | null = null
let statusChart: Chart | null = null

const filters = reactive({
  search: '',
  status: '',
  dateRange: [],
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

async function loadRuns() {
  if (!workspaceStore.currentWorkspaceId) return
  loading.value = true
  try {
    const data = await runAPI.list({
      workspaceId: workspaceStore.currentWorkspaceIdOrThrow,
      page: pagination.page,
      pageSize: pagination.pageSize,
    })
    runs.value = data.list
    pagination.total = data.total
  } catch (error) {
    ElMessage.error('加载运行记录失败')
  } finally {
    loading.value = false
  }
}

async function loadStats() {
  if (!workspaceStore.currentWorkspaceId) return
  try {
    const data = await runAPI.getStats({
      workspaceId: workspaceStore.currentWorkspaceIdOrThrow,
    })
    runStats.value = data

    stats.value = [
      { title: '总运行次数', value: data.totalRuns, change: 0 },
      { title: '成功率', value: `${data.successRate}%`, change: 0 },
      { title: '平均耗时', value: formatDuration(data.avgDuration), change: 0 },
      { title: '今日运行', value: data.todayRuns, change: 0 },
    ]

    await nextTick()
    renderTrendChart(data.trend ?? [])
    renderStatusChart(data)
  } catch (error) {
    console.error('加载统计失败', error)
  }
}

function renderTrendChart(trend: Array<{ date: string; success: number; failed: number }>) {
  if (!trendChartRef.value) return

  if (trendChart) {
    trendChart.destroy()
  }

  const labels = trend.map((item) => item.date)
  trendChart = new Chart(trendChartRef.value, {
    type: 'line',
    data: {
      labels,
      datasets: [
        {
          label: '成功',
          data: trend.map((item) => item.success),
          borderColor: '#67c23a',
          backgroundColor: 'rgba(103, 194, 58, 0.1)',
          fill: true,
          tension: 0.4,
        },
        {
          label: '失败',
          data: trend.map((item) => item.failed),
          borderColor: '#f56c6c',
          backgroundColor: 'rgba(245, 108, 108, 0.1)',
          fill: true,
          tension: 0.4,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: { position: 'top' },
      },
      scales: {
        y: { beginAtZero: true, ticks: { stepSize: 1 } },
      },
    },
  })
}

function renderStatusChart(stats: RunStats) {
  if (!statusChartRef.value) return

  if (statusChart) {
    statusChart.destroy()
  }

  const failed = stats.totalRuns - stats.todayRuns > 0
    ? Math.round((1 - stats.successRate / 100) * stats.totalRuns)
    : 0
  const success = stats.totalRuns - failed

  statusChart = new Chart(statusChartRef.value, {
    type: 'doughnut',
    data: {
      labels: ['成功', '失败'],
      datasets: [
        {
          data: [success, failed],
          backgroundColor: ['#67c23a', '#f56c6c'],
          borderWidth: 0,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: { position: 'bottom' },
      },
      cutout: '65%',
    },
  })
}

onBeforeUnmount(() => {
  trendChart?.destroy()
  statusChart?.destroy()
})

function handleSearch() {
  pagination.page = 1
  loadRuns()
}

function handleView(id: string) {
  router.push(runDetailPath(id))
}

async function handleRetry(id: string) {
  try {
    await runAPI.retry(id)
    ElMessage.success('已提交重试')
    loadRuns()
  } catch (error) {
    ElMessage.error('重试失败')
  }
}

function handleSizeChange() {
  loadRuns()
}

function handleCurrentChange() {
  loadRuns()
}

onMounted(() => {
  loadRuns()
  loadStats()
})

watch(() => workspaceStore.workspaces.length, (len) => {
  if (len > 0 && workspaceStore.currentWorkspaceId) {
    loadRuns()
    loadStats()
  }
})
</script>

<style scoped>
.run-list {
  padding: 20px;
}

.header {
  margin-bottom: 30px;
}

.stats-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 30px;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #303133;
}

.stat-change {
  font-size: 14px;
  margin-top: 8px;
}

.stat-change.up {
  color: #67c23a;
}

.stat-change.down {
  color: #f56c6c;
}

.filters {
  margin-bottom: 30px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 4px;
}

.charts {
  margin-bottom: 30px;
}

.chart-card {
  height: 300px;
}

.chart-container {
  height: calc(100% - 60px);
  position: relative;
}

.chart-container--small {
  display: flex;
  align-items: center;
  justify-content: center;
}

.run-id-link {
  color: #409eff;
  text-decoration: none;
}

.run-id-link:hover {
  text-decoration: underline;
}

.pagination {
  margin-top: 30px;
  display: flex;
  justify-content: flex-end;
}
</style>
