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
            icon="PlayIcon"
            color="primary"
          />
          
          <m-stat-card
            title="Total Runs"
            :value="statistics.totalRuns"
            :change="8"
            change-type="increase"
            icon="ArrowPathIcon"
            color="secondary"
          />
          
          <m-stat-card
            title="Success Rate"
            :value="`${statistics.successRate}%`"
            :change="-2"
            change-type="decrease"
            icon="CheckCircleIcon"
            color="success"
          />
          
          <m-stat-card
            title="Avg. Execution Time"
            :value="`${statistics.avgExecutionTime}s`"
            :change="-15"
            change-type="decrease"
            icon="ClockIcon"
            color="info"
          />
        </div>
        
        <!-- 最近活动 -->
        <div class="recent-activity">
          <div class="section-header">
            <h2 class="section-title">Recent Activity</h2>
            <router-link to="/activity" class="view-all-link">
              View All
            </router-link>
          </div>
          
          <div class="activity-list">
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
          
          <div class="runs-table-wrapper">
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
        </div>
        
        <!-- 资源使用情况 -->
        <div class="resource-usage">
          <div class="section-header">
            <h2 class="section-title">Resource Usage</h2>
          </div>
          
          <div class="usage-metrics">
            <div class="usage-metric">
              <div class="metric-header">
                <h4 class="metric-title">CPU Usage</h4>
                <span class="metric-value">{{ resourceUsage.cpu }}%</span>
              </div>
              <div class="metric-progress">
                <div class="progress-bar">
                  <div
                    class="progress-fill"
                    :class="{ 'warning': resourceUsage.cpu > 80, 'danger': resourceUsage.cpu > 95 }"
                    :style="{ width: `${resourceUsage.cpu}%` }"
                  />
                </div>
              </div>
            </div>
            
            <div class="usage-metric">
              <div class="metric-header">
                <h4 class="metric-title">Memory Usage</h4>
                <span class="metric-value">{{ resourceUsage.memory }}%</span>
              </div>
              <div class="metric-progress">
                <div class="progress-bar">
                  <div
                    class="progress-fill"
                    :class="{ 'warning': resourceUsage.memory > 80, 'danger': resourceUsage.memory > 95 }"
                    :style="{ width: `${resourceUsage.memory}%` }"
                  />
                </div>
              </div>
            </div>
            
            <div class="usage-metric">
              <div class="metric-header">
                <h4 class="metric-title">Database Connections</h4>
                <span class="metric-value">{{ resourceUsage.dbConnections }}/{{ resourceUsage.dbMaxConnections }}</span>
              </div>
              <div class="metric-progress">
                <div class="progress-bar">
                  <div
                    class="progress-fill"
                    :class="{ 'warning': (resourceUsage.dbConnections / resourceUsage.dbMaxConnections * 100) > 80, 'danger': (resourceUsage.dbConnections / resourceUsage.dbMaxConnections * 100) > 95 }"
                    :style="{ width: `${(resourceUsage.dbConnections / resourceUsage.dbMaxConnections * 100)}%` }"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- 即将到期的触发器 -->
    <div class="upcoming-triggers">
      <div class="section-header">
        <h2 class="section-title">Upcoming Triggers</h2>
        <router-link to="/triggers" class="view-all-link">
          View All
        </router-link>
      </div>
      
      <div class="triggers-list">
        <div
          v-for="trigger in upcomingTriggers"
          :key="trigger.id"
          class="trigger-card"
        >
          <div class="trigger-header">
            <div class="trigger-icon-wrapper">
              <component :is="getTriggerIcon(trigger.type)" class="h-5 w-5" />
            </div>
            <div class="trigger-info">
              <h4 class="trigger-name">{{ trigger.name }}</h4>
              <p class="trigger-workflow">{{ trigger.workflow }}</p>
            </div>
            <el-tag
              :type="getTriggerStatusType(trigger.type)"
              size="small"
              class="trigger-type"
            >
              {{ trigger.type }}
            </el-tag>
          </div>
          
          <div class="trigger-details">
            <div class="trigger-detail">
              <CalendarIcon class="h-4 w-4" />
              <span>{{ trigger.schedule }}</span>
            </div>
            <div class="trigger-detail">
              <ClockIcon class="h-4 w-4" />
              <span>Next run: {{ trigger.nextRun }}</span>
            </div>
          </div>
          
          <div class="trigger-actions">
            <m-button variant="text" size="xs" @click="viewTrigger(trigger)">
              View
            </m-button>
            <m-button variant="text" size="xs" @click="editTrigger(trigger)">
              Edit
            </m-button>
          </div>
        </div>
      </div>
      
      <div v-if="upcomingTriggers.length === 0" class="empty-triggers">
        <div class="empty-triggers-content">
          <CalendarIcon class="empty-triggers-icon" />
          <p class="empty-triggers-title">No upcoming triggers</p>
          <p class="empty-triggers-description">
            Create triggers to schedule your workflow executions.
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import MButton from '@/components/ui/MButton.vue'
import MStatCard from '@/components/ui/MStatCard.vue'

// Icons (imported from Element Plus or custom)
import {
  PlusIcon,
  ArrowsRightLeftIcon,
  TemplateIcon,
  ChartBarIcon,
  PuzzleIcon,
  PlayIcon,
  ArrowPathIcon,
  CheckCircleIcon,
  ClockIcon,
  CalendarIcon,
  ChartNoDataIcon,
} from '@/components/icons'

const router = useRouter()
const authStore = useAuthStore()

// Data
const user = computed(() => ({
  displayName: authStore.displayName || 'Guest',
}))

const statistics = ref({
  activeWorkflows: 24,
  totalRuns: 1568,
  successRate: 94,
  avgExecutionTime: 2.3,
})

const recentActivity = ref([
  { id: 1, type: 'success', message: 'Workflow "Order Processing" completed successfully', time: '5 min ago', status: 'completed' },
  { id: 2, type: 'warning', message: 'Trigger "Daily Report" scheduled for tomorrow', time: '1 hour ago', status: 'scheduled' },
  { id: 3, type: 'info', message: 'New user "john.doe@example.com" registered', time: '3 hours ago', status: 'created' },
  { id: 4, type: 'error', message: 'Workflow "Data Sync" failed with timeout error', time: '5 hours ago', status: 'failed' },
  { id: 5, type: 'success', message: 'API integration "Shopify" connected successfully', time: '1 day ago', status: 'connected' },
])

const workflowRuns = ref([
  { id: 1, name: 'Order Processing', type: 'workflow', status: 'success', started: '5 min ago', duration: '1.2s' },
  { id: 2, name: 'Daily Report', type: 'timer', status: 'running', started: '2 min ago', duration: '0:45' },
  { id: 3, name: 'Data Sync', type: 'sync', status: 'failed', started: '5 hours ago', duration: '30.5s' },
  { id: 4, name: 'Email Campaign', type: 'marketing', status: 'success', started: '1 day ago', duration: '5.3s' },
  { id: 5, name: 'User Onboarding', type: 'workflow', status: 'success', started: '1 day ago', duration: '2.1s' },
])

const resourceUsage = ref({
  cpu: 45,
  memory: 68,
  dbConnections: 12,
  dbMaxConnections: 20,
})

const upcomingTriggers = ref([
  { id: 1, name: 'Daily Sales Report', type: 'timer', workflow: 'Report Generator', schedule: 'Daily at 09:00', nextRun: 'in 8 hours' },
  { id: 2, name: 'Weekly Backup', type: 'timer', workflow: 'Database Backup', schedule: 'Weekly on Monday', nextRun: 'in 1 day' },
  { id: 3, name: 'Order Webhook', type: 'webhook', workflow: 'Order Processing', schedule: 'On order created', nextRun: 'Real-time' },
  { id: 4, name: 'Inventory Sync', type: 'mq', workflow: 'Stock Sync', schedule: 'Every 2 hours', nextRun: 'in 1 hour' },
])

// State
const workflowFilter = ref('all')

// Computed
const filteredWorkflowRuns = computed(() => {
  if (workflowFilter.value === 'all') {
    return workflowRuns.value
  }
  return workflowRuns.value.filter(run => run.status === workflowFilter.value)
})

// Methods
const navigateTo = (section: string) => {
  router.push(`/${section}`)
}

const createWorkflow = () => {
  router.push('/workflows/create')
}

const viewRunDetails = (run: any) => {
  console.log('View run details:', run)
}

const retryRun = (run: any) => {
  console.log('Retry run:', run)
}

const viewTrigger = (trigger: any) => {
  console.log('View trigger:', trigger)
}

const editTrigger = (trigger: any) => {
  console.log('Edit trigger:', trigger)
}

const getActivityIcon = (type: string) => {
  const icons = {
    success: 'CheckCircleIcon',
    error: 'XCircleIcon',
    warning: 'ExclamationTriangleIcon',
    info: 'InformationCircleIcon',
  }
  return icons[type] || 'InformationCircleIcon'
}

const getWorkflowIcon = (type: string) => {
  const icons = {
    workflow: 'ArrowsRightLeftIcon',
    timer: 'ClockIcon',
    sync: 'ArrowPathIcon',
    marketing: 'EnvelopeIcon',
  }
  return icons[type] || 'ArrowsRightLeftIcon'
}

const getTriggerIcon = (type: string) => {
  const icons = {
    timer: 'ClockIcon',
    webhook: 'LinkIcon',
    mq: 'QueueListIcon',
  }
  return icons[type] || 'ClockIcon'
}

const getStatusType = (status: string) => {
  const types = {
    completed: 'success',
    scheduled: 'info',
    created: 'info',
    failed: 'danger',
    connected: 'success',
  }
  return types[status] || 'info'
}

const getRunStatusType = (status: string) => {
  const types = {
    success: 'success',
    running: 'warning',
    failed: 'danger',
  }
  return types[status] || 'info'
}

const getTriggerStatusType = (type: string) => {
  const types = {
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

/* 资源使用情况 */
.resource-usage {
  @apply bg-white border border-gray-200 rounded-lg p-5;
}

.usage-metrics {
  @apply space-y-4;
}

.usage-metric {
  @apply space-y-2;
}

.metric-header {
  @apply flex items-center justify-between;
}

.metric-title {
  @apply text-sm font-medium text-gray-700;
}

.metric-value {
  @apply text-sm font-semibold text-gray-900;
}

.progress-bar {
  @apply h-2 bg-gray-200 rounded-full overflow-hidden;
}

.progress-fill {
  @apply h-full bg-primary-500 rounded-full transition-all duration-300;
}

.progress-fill.warning {
  @apply bg-warning-500;
}

.progress-fill.danger {
  @apply bg-danger-500;
}

/* 即将到期的触发器 */
.upcoming-triggers {
  @apply bg-white border border-gray-200 rounded-lg p-5;
}

.triggers-list {
  @apply grid grid-cols-1 md:grid-cols-2 gap-4;
}

.trigger-card {
  @apply border border-gray-200 rounded-lg p-4 hover:border-primary-300 hover:shadow-sm transition-all duration-200;
}

.trigger-header {
  @apply flex items-start justify-between mb-3;
}

.trigger-icon-wrapper {
  @apply h-10 w-10 rounded-lg bg-primary-50 flex items-center justify-center mr-3;
}

.trigger-icon-wrapper svg {
  @apply text-primary-600;
}

.trigger-info {
  @apply flex-1 min-w-0;
}

.trigger-name {
  @apply text-sm font-semibold text-gray-900 truncate;
}

.trigger-workflow {
  @apply text-xs text-gray-500 truncate;
}

.trigger-type {
  @apply flex-shrink-0;
}

.trigger-details {
  @apply space-y-2 mb-3;
}

.trigger-detail {
  @apply flex items-center gap-2 text-xs text-gray-600;
}

.trigger-actions {
  @apply flex gap-2;
}

.empty-triggers {
  @apply py-8 text-center;
}

.empty-triggers-content {
  @apply space-y-2;
}

.empty-triggers-icon {
  @apply h-10 w-10 mx-auto text-gray-400;
}

.empty-triggers-title {
  @apply text-sm font-medium text-gray-900;
}

.empty-triggers-description {
  @apply text-sm text-gray-500;
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