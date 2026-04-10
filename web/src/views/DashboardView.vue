<template>
  <div class="dashboard">
    <!-- 页面标题 -->
    <div class="dashboard-header">
      <div>
        <h1 class="dashboard-title">
          Welcome back, {{ user.displayName }}
        </h1>
        <p class="dashboard-subtitle">
          Here's what's happening with your workflows today.
        </p>
      </div>

      <div class="dashboard-actions">
        <m-button variant="primary" size="md" @click="createWorkflow">
          <template #icon>
            <PlusIcon class="h-4 w-4" />
          </template>
          New Workflow
        </m-button>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="isLoading" class="loading-container">
      <el-icon class="is-loading"><Loading /></el-icon>
      <span>Loading dashboard data...</span>
    </div>

    <!-- 主要内容区域 -->
    <template v-else>
      <!-- 快捷操作卡片 -->
      <div class="quick-actions">
        <div class="quick-action-card" @click="navigateTo('workflows')">
          <div class="action-icon-wrapper">
            <ArrowsRightLeftIcon class="action-icon h-6 w-6" />
          </div>
          <h3 class="action-title">Workflows</h3>
          <p class="action-description">Manage your automation workflows</p>
        </div>

        <div class="quick-action-card" @click="navigateTo('templates')">
          <div class="action-icon-wrapper">
            <TemplateIcon class="action-icon h-6 w-6" />
          </div>
          <h3 class="action-title">Templates</h3>
          <p class="action-description">Start from pre-built templates</p>
        </div>

        <div class="quick-action-card" @click="navigateTo('analytics')">
          <div class="action-icon-wrapper">
            <ChartBarIcon class="action-icon h-6 w-6" />
          </div>
          <h3 class="action-title">Analytics</h3>
          <p class="action-description">View performance insights</p>
        </div>

        <div class="quick-action-card" @click="navigateTo('integrations')">
          <div class="action-icon-wrapper">
            <PuzzleIcon class="action-icon h-6 w-6" />
          </div>
          <h3 class="action-title">Integrations</h3>
          <p class="action-description">Connect your tools</p>
        </div>
      </div>

      <!-- 主要内容区域 -->
      <div class="dashboard-content">
        <!-- 左栏：统计数据 -->
        <div class="dashboard-left">
          <!-- 统计卡片网格 -->
          <div class="statistics-grid">
            <m-stat-card
              title="Active Workflows"
              :value="statistics.activeWorkflows"
              :change="12"
              change-type="increase"
              :icon="PlayIcon"
              color="primary"
            />

            <m-stat-card
              title="Total Runs"
              :value="statistics.totalRuns"
              :change="8"
              change-type="increase"
              :icon="ArrowPathIcon"
              color="secondary"
            />

            <m-stat-card
              title="Success Rate"
              :value="`${statistics.successRate}%`"
              :change="-2"
              change-type="decrease"
              :icon="CheckCircleIcon"
              color="success"
            />

            <m-stat-card
              title="Avg. Execution Time"
              :value="`${statistics.avgExecutionTime}s`"
              :change="-15"
              change-type="decrease"
              :icon="ClockIcon"
              color="info"
            />
          </div>

          <!-- 最近活动 -->
          <div class="recent-activity">
            <div class="section-header">
              <h2 class="section-title">Recent Activity</h2>
              <router-link to="/runs" class="view-all-link">
                View All
              </router-link>
            </div>

            <div v-if="recentActivity.length === 0" class="empty-activity">
              <p class="text-gray-500">No recent activity</p>
            </div>

            <div v-else class="activity-list">
              <div
                v-for="activity in recentActivity"
                :key="activity.id"
                class="activity-item"
              >
                <div class="activity-icon-wrapper" :class="`type-${activity.type}`">
                  <component :is="getActivityIcon(activity.type)" class="h-4 w-4" />
                </div>
                <div class="activity-content">
                  <p class="activity-message">{{ activity.message }}</p>
                  <span class="activity-time">{{ activity.time }}</span>
                </div>
                <el-tag
                  v-if="activity.status"
                  :type="getStatusType(activity.status)"
                  size="small"
                  class="activity-status"
                >
                  {{ activity.status }}
                </el-tag>
              </div>
            </div>
          </div>
        </div>

        <!-- 右栏：工作流最近运行 -->
        <div class="dashboard-right">
          <div class="workflow-runs">
            <div class="section-header">
              <h2 class="section-title">Recent Workflow Runs</h2>
              <div class="filter-dropdown">
                <el-select
                  v-model="workflowFilter"
                  size="small"
                  placeholder="Filter by status"
                  class="w-full"
                >
                  <el-option label="All" value="all" />
                  <el-option label="Success" value="success" />
                  <el-option label="Failed" value="failed" />
                  <el-option label="Running" value="running" />
                </el-select>
              </div>
            </div>

            <div v-if="filteredWorkflowRuns.length === 0" class="empty-state">
              <div class="empty-state-content">
                <ChartNoDataIcon class="empty-state-icon" />
                <p class="empty-state-title">No workflow runs found</p>
                <p class="empty-state-description">
                  {{ workflowFilter === 'all' ? 'Start by creating your first workflow.' : `No ${workflowFilter} runs found.` }}
                </p>
                <m-button variant="primary" size="sm" @click="createWorkflow">
                  Create Workflow
                </m-button>
              </div>
            </div>

            <div v-else class="runs-table-wrapper">
              <table class="runs-table">
                <thead>
                  <tr>
                    <th class="text-left">Workflow</th>
                    <th class="text-left">Status</th>
                    <th class="text-left">Started</th>
                    <th class="text-left">Duration</th>
                    <th class="text-right">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="run in filteredWorkflowRuns"
                    :key="run.id"
                    class="run-row"
                  >
                    <td class="run-name">
                      <div class="flex items-center gap-2">
                        <div class="run-icon-wrapper">
                          <component :is="getWorkflowIcon(run.type)" class="h-4 w-4" />
                        </div>
                        <span class="truncate">{{ run.name }}</span>
                      </div>
                    </td>
                    <td>
                      <el-tag
                        :type="getRunStatusType(run.status)"
                        size="small"
                        class="run-status"
                        :class="`status-${run.status}`"
                      >
                        {{ run.status }}
                      </el-tag>
                    </td>
                    <td>
                      {{ run.started }}
                    </td>
                    <td>
                      {{ run.duration }}
                    </td>
                    <td class="run-actions">
                      <div class="flex items-center justify-end gap-1">
                        <m-button
                          variant="text"
                          size="xs"
                          @click="viewRunDetails(run)"
                        >
                          View
                        </m-button>
                        <m-button
                          v-if="run.status === 'failed'"
                          variant="text"
                          size="xs"
                          @click="retryRun(run)"
                        >
                          Retry
                        </m-button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- 触发器列表 -->
          <div class="resource-usage">
            <div class="section-header">
              <h2 class="section-title">Active Triggers</h2>
              <router-link to="/triggers" class="view-all-link">
                View All
              </router-link>
            </div>

            <div v-if="upcomingTriggers.length === 0" class="empty-triggers">
              <p class="text-gray-500 text-sm">No active triggers</p>
            </div>

            <div v-else class="triggers-list">
              <div
                v-for="trigger in upcomingTriggers"
                :key="trigger.id"
                class="trigger-item"
              >
                <div class="trigger-icon-wrapper" :class="`type-${trigger.type}`">
                  <component :is="getTriggerIcon(trigger.type)" class="h-4 w-4" />
                </div>
                <div class="trigger-info">
                  <h4 class="trigger-name">{{ trigger.name }}</h4>
                  <p class="trigger-schedule">{{ trigger.schedule }}</p>
                </div>
                <el-tag
                  :type="getTriggerStatusType(trigger.type)"
                  size="small"
                >
                  {{ trigger.type }}
                </el-tag>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Loading } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { useDashboard } from '@/composables/useDashboard'
import { runAPI } from '@/api/runs'
import MButton from '@/components/ui/MButton.vue'
import MStatCard from '@/components/ui/MStatCard.vue'

// Icons
import {
  PlusIcon,
  ArrowsRightLeftIcon,
  TemplateIcon,
  ChartBarIcon,
  PuzzleIcon,
  ClockIcon,
  ChartNoDataIcon,
  QueueListIcon,
  LinkIcon,
  CheckCircleIcon,
  XCircleIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
  ArrowPathIcon,
  EnvelopeIcon,
  PlayIcon,
} from '@/components/icons'

const router = useRouter()
const authStore = useAuthStore()
const {
  isLoading,
  statistics,
  recentActivity,
  workflowRuns,
  upcomingTriggers,
  loadDashboardData,
} = useDashboard()

// 用户信息
const user = computed(() => ({
  displayName: authStore.displayName || 'Guest',
}))

// 过滤器
const workflowFilter = ref('all')

// 过滤后的工作流运行
const filteredWorkflowRuns = computed(() => {
  if (workflowFilter.value === 'all') {
    return workflowRuns.value
  }
  return workflowRuns.value.filter(run => run.status === workflowFilter.value)
})

// 加载数据
onMounted(() => {
  loadDashboardData()
})

// 导航方法 - key 到实际路由的映射
const routeMap: Record<string, string> = {
  workflows: '/flows',
  templates: '/flows',
  analytics: '/runs',
  integrations: '/resources',
}

const navigateTo = (section: string) => {
  router.push(routeMap[section] || `/${section}`)
}

const createWorkflow = () => {
  router.push('/flows/new')
}

const viewRunDetails = async (run: any) => {
  router.push(`/runs/${run.id}`)
}

const retryRun = async (run: any) => {
  try {
    await runAPI.retry(run.id)
    // 重新加载数据
    loadDashboardData()
  } catch (error) {
    console.error('Failed to retry run:', error)
  }
}

// 图标映射
const getActivityIcon = (type: string) => {
  const icons: Record<string, any> = {
    success: CheckCircleIcon,
    error: XCircleIcon,
    warning: ExclamationTriangleIcon,
    info: InformationCircleIcon,
  }
  return icons[type] || InformationCircleIcon
}

const getWorkflowIcon = (type: string) => {
  const icons: Record<string, any> = {
    workflow: ArrowsRightLeftIcon,
    timer: ClockIcon,
    sync: ArrowPathIcon,
    marketing: EnvelopeIcon,
  }
  return icons[type] || ArrowsRightLeftIcon
}

const getTriggerIcon = (type: string) => {
  const icons: Record<string, any> = {
    timer: ClockIcon,
    webhook: LinkIcon,
    mq: QueueListIcon,
  }
  return icons[type] || ClockIcon
}

const getStatusType = (status: string) => {
  const types: Record<string, string> = {
    completed: 'success',
    scheduled: 'info',
    created: 'info',
    failed: 'danger',
    connected: 'success',
    running: 'warning',
  }
  return types[status] || 'info'
}

const getRunStatusType = (status: string) => {
  const types: Record<string, string> = {
    success: 'success',
    running: 'warning',
    failed: 'danger',
  }
  return types[status] || 'info'
}

const getTriggerStatusType = (type: string) => {
  const types: Record<string, string> = {
    timer: 'warning',
    webhook: 'primary',
    mq: 'success',
  }
  return types[type] || 'info'
}
</script>

<style scoped>
.dashboard {
  @apply space-y-6;
}

/* Dashboard 头部样式 */
.dashboard-header {
  @apply flex flex-col sm:flex-row sm:items-center justify-between gap-4;
}

.dashboard-title {
  @apply text-2xl font-bold text-gray-900;
}

.dashboard-subtitle {
  @apply text-gray-600 mt-1;
}

.dashboard-actions {
  @apply flex-shrink-0;
}

/* 加载状态 */
.loading-container {
  @apply flex items-center justify-center gap-2 py-12 text-gray-500;
}

/* 快捷操作卡片 */
.quick-actions {
  @apply grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4;
}

.quick-action-card {
  @apply bg-white border border-gray-200 rounded-lg p-5 cursor-pointer transition-all duration-200 hover:shadow-md hover:border-primary-200 hover:translate-y-[-2px];
  @apply flex flex-col items-center text-center;
}

.action-icon-wrapper {
  @apply h-12 w-12 rounded-full bg-primary-50 flex items-center justify-center mb-3;
}

.action-icon {
  @apply text-primary-600;
}

.action-title {
  @apply text-base font-semibold text-gray-900 mb-1;
}

.action-description {
  @apply text-sm text-gray-500;
}

/* 主要内容区域 */
.dashboard-content {
  @apply grid grid-cols-1 lg:grid-cols-3 gap-6;
}

.dashboard-left {
  @apply lg:col-span-2 space-y-6;
}

.dashboard-right {
  @apply space-y-6;
}

/* 统计卡片网格 */
.statistics-grid {
  @apply grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4;
}

/* 章节头部 */
.section-header {
  @apply flex items-center justify-between mb-4;
}

.section-title {
  @apply text-lg font-semibold text-gray-900;
}

.view-all-link {
  @apply text-sm text-primary-600 hover:text-primary-800 no-underline;
}

/* 最近活动 */
.activity-list {
  @apply bg-white border border-gray-200 rounded-lg divide-y divide-gray-200;
}

.activity-item {
  @apply p-4 flex items-start gap-3;
}

.activity-icon-wrapper {
  @apply h-8 w-8 rounded-full flex items-center justify-center flex-shrink-0;
}

.activity-icon-wrapper.type-success {
  @apply bg-green-50 text-green-600;
}

.activity-icon-wrapper.type-warning {
  @apply bg-yellow-50 text-yellow-600;
}

.activity-icon-wrapper.type-info {
  @apply bg-blue-50 text-blue-600;
}

.activity-icon-wrapper.type-error {
  @apply bg-red-50 text-red-600;
}

.activity-content {
  @apply flex-1 min-w-0;
}

.activity-message {
  @apply text-sm text-gray-900;
}

.activity-time {
  @apply text-xs text-gray-500 mt-1 block;
}

.activity-status {
  @apply flex-shrink-0;
}

.empty-activity {
  @apply bg-white border border-gray-200 rounded-lg p-8 text-center;
}

/* 工作流运行表格 */
.workflow-runs {
  @apply bg-white border border-gray-200 rounded-lg p-5;
}

.filter-dropdown {
  @apply w-32;
}

.runs-table-wrapper {
  @apply overflow-x-auto;
}

.runs-table {
  @apply w-full text-sm;
}

.runs-table th {
  @apply px-4 py-3 text-xs font-semibold text-gray-600 uppercase tracking-wide border-b border-gray-200;
}

.runs-table td {
  @apply px-4 py-3 border-b border-gray-200;
}

.run-row {
  @apply hover:bg-gray-50;
}

.run-row:last-child td {
  @apply border-b-0;
}

.run-name {
  @apply max-w-[200px];
}

.run-icon-wrapper {
  @apply h-6 w-6 rounded bg-gray-100 flex items-center justify-center text-gray-600;
}

.run-status {
  @apply capitalize;
}

.run-status.status-success {
  @apply bg-green-50 text-green-700 border-green-100;
}

.run-status.status-running {
  @apply bg-yellow-50 text-yellow-700 border-yellow-100;
}

.run-status.status-failed {
  @apply bg-red-50 text-red-700 border-red-100;
}

.run-actions {
  @apply whitespace-nowrap;
}

.empty-state {
  @apply py-10 text-center;
}

.empty-state-content {
  @apply space-y-3;
}

.empty-state-icon {
  @apply h-12 w-12 mx-auto text-gray-400;
}

.empty-state-title {
  @apply text-sm font-medium text-gray-900;
}

.empty-state-description {
  @apply text-sm text-gray-500;
}

/* 触发器列表 */
.resource-usage {
  @apply bg-white border border-gray-200 rounded-lg p-5;
}

.triggers-list {
  @apply space-y-3;
}

.trigger-item {
  @apply flex items-center gap-3 p-3 border border-gray-100 rounded-lg hover:bg-gray-50 transition-colors;
}

.trigger-icon-wrapper {
  @apply h-8 w-8 rounded-lg flex items-center justify-center flex-shrink-0;
}

.trigger-icon-wrapper.type-timer {
  @apply bg-warning-50 text-warning-600;
}

.trigger-icon-wrapper.type-webhook {
  @apply bg-primary-50 text-primary-600;
}

.trigger-icon-wrapper.type-mq {
  @apply bg-success-50 text-success-600;
}

.trigger-info {
  @apply flex-1 min-w-0;
}

.trigger-name {
  @apply text-sm font-medium text-gray-900 truncate;
}

.trigger-schedule {
  @apply text-xs text-gray-500 truncate;
}

.empty-triggers {
  @apply py-4 text-center;
}

/* 响应式调整 */
@media (max-width: 640px) {
  .dashboard-content {
    @apply gap-4;
  }

  .statistics-grid {
    @apply gap-3;
  }

  .quick-actions {
    @apply gap-3;
  }
}
</style>
