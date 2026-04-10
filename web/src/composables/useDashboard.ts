import { ref, computed } from 'vue'
import { runAPI } from '@/api/runs'
import { flowAPI } from '@/api/flows'
import { triggerAPI } from '@/api/triggers'
import { useWorkspaceStore } from '@/stores/workspace'
import type { Run } from '@/types/run'
import type { Flow } from '@/types/flow'
import type { Trigger } from '@/types/trigger'

// 格式化时间显示
function formatTimeAgo(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = Math.floor((now.getTime() - date.getTime()) / 1000)

  if (diff < 60) return 'just now'
  if (diff < 3600) return `${Math.floor(diff / 60)} min ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)} hour${Math.floor(diff / 3600) > 1 ? 's' : ''} ago`
  if (diff < 604800) return `${Math.floor(diff / 86400)} day${Math.floor(diff / 86400) > 1 ? 's' : ''} ago`
  return date.toLocaleDateString()
}

// 格式化执行时长
function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`
  const seconds = (ms / 1000).toFixed(1)
  return `${seconds}s`
}

export function useDashboard() {
  const workspaceStore = useWorkspaceStore()

  // 加载状态
  const loading = ref(false)
  const statsLoading = ref(false)
  const runsLoading = ref(false)
  const triggersLoading = ref(false)
  const flowsLoading = ref(false)

  // 统计数据
  const statistics = ref({
    activeWorkflows: 0,
    totalRuns: 0,
    successRate: 0,
    avgExecutionTime: 0,
  })

  // 最近活动
  const recentActivity = ref<Array<{
    id: string
    type: 'success' | 'warning' | 'info' | 'error'
    message: string
    time: string
    status: string
  }>>([])

  // 工作流运行记录
  const workflowRuns = ref<Array<{
    id: string
    name: string
    type: string
    status: 'success' | 'failed' | 'running' | 'timeout'
    started: string
    duration: string
    triggerId?: string
    flowId?: string
  }>>([])

  // 即将到期的触发器
  const upcomingTriggers = ref<Array<{
    id: string
    name: string
    type: 'timer' | 'webhook' | 'mq' | string
    workflow: string
    schedule: string
    nextRun: string
    status: string
  }>>([])

  // 活跃工作流数量
  const activeWorkflows = ref<Flow[]>([])

  // 加载统计数据
  async function loadStats() {
    if (!workspaceStore.currentWorkspaceId) return

    statsLoading.value = true
    try {
      const stats = await runAPI.getStats({ workspaceId: workspaceStore.currentWorkspaceId })
      statistics.value = {
        activeWorkflows: activeWorkflows.value.length,
        totalRuns: stats.totalRuns,
        successRate: Math.round(stats.successRate),
        avgExecutionTime: stats.avgDuration > 0 ? stats.avgDuration / 1000 : 0,
      }
    } catch (error) {
      console.error('Failed to load stats:', error)
    } finally {
      statsLoading.value = false
    }
  }

  // 加载最近运行记录
  async function loadRecentRuns() {
    if (!workspaceStore.currentWorkspaceId) return

    runsLoading.value = true
    try {
      const result = await runAPI.list({
        workspaceId: workspaceStore.currentWorkspaceId,
        page: 1,
        pageSize: 10,
      })

      // 转换运行记录
      workflowRuns.value = result.list.map((run: Run) => ({
        id: run.id,
        name: run.flowName,
        type: 'workflow',
        status: run.status,
        started: formatTimeAgo(run.startedAt),
        duration: formatDuration(run.durationMs),
        triggerId: run.triggerId,
        flowId: run.flowId,
      }))

      // 生成最近活动
      recentActivity.value = result.list.slice(0, 5).map((run: Run) => {
        let type: 'success' | 'warning' | 'info' | 'error' = 'info'
        let status = 'unknown'

        switch (run.status) {
          case 'success':
            type = 'success'
            status = 'completed'
            break
          case 'failed':
            type = 'error'
            status = 'failed'
            break
          case 'running':
            type = 'info'
            status = 'running'
            break
          case 'timeout':
            type = 'warning'
            status = 'timeout'
            break
        }

return {
        id: run.id,
        type,
        message: `Workflow "${run.flowName}" ${run.status === 'success' ? 'completed successfully' : run.status === 'failed' ? 'failed' : run.status === 'timeout' ? 'timed out' : 'is running'}`,
        time: formatTimeAgo(run.startedAt),
        status,
      }
      })
    } catch (error) {
      console.error('Failed to load runs:', error)
    } finally {
      runsLoading.value = false
    }
  }

  // 加载触发器
  async function loadTriggers() {
    if (!workspaceStore.currentWorkspaceId) return

    triggersLoading.value = true
    try {
      const response = await triggerAPI.list({
        workspaceId: workspaceStore.currentWorkspaceId,
      })

      upcomingTriggers.value = response.list
        .filter((t: Trigger) => t.status === 'active')
        .slice(0, 4)
        .map((trigger: Trigger) => {
          // 格式化触发器信息
          let schedule = ''
          let nextRun = 'Real-time'

          switch (trigger.type) {
            case 'timer':
              const timerConfig = trigger.config as { cron?: string; interval?: number }
              if (timerConfig.cron) {
                schedule = `Cron: ${timerConfig.cron}`
                nextRun = 'Scheduled'
              } else if (timerConfig.interval) {
                const hours = Math.floor(timerConfig.interval / 3600)
                const minutes = Math.floor((timerConfig.interval % 3600) / 60)
                schedule = hours > 0 ? `Every ${hours}h` : `Every ${minutes}m`
                nextRun = 'Scheduled'
              }
              break
            case 'restful':
              schedule = 'On demand'
              nextRun = 'Real-time'
              break
            case 'rabbitmq':
            case 'kafka':
              schedule = `Queue: ${(trigger.config as { queue?: string; topic?: string }).queue || (trigger.config as { queue?: string; topic?: string }).topic || 'default'}`
              nextRun = 'Real-time'
              break
            default:
              schedule = trigger.type
          }

          // 映射触发器类型到前端显示
          const typeMap: Record<string, string> = {
            restful: 'webhook',
            timer: 'timer',
            rabbitmq: 'mq',
            kafka: 'mq',
          }

          return {
            id: trigger.id,
            name: trigger.name,
            type: typeMap[trigger.type] || trigger.type,
            workflow: trigger.flowId,
            schedule,
            nextRun,
            status: trigger.status,
          }
        })
    } catch (error) {
      console.error('Failed to load triggers:', error)
    } finally {
      triggersLoading.value = false
    }
  }

  // 加载工作流
  async function loadWorkflows() {
    if (!workspaceStore.currentWorkspaceId) return

    flowsLoading.value = true
    try {
      const result = await flowAPI.list({
        workspaceId: workspaceStore.currentWorkspaceId,
        status: 'active',
      })

      activeWorkflows.value = result.list || []
      statistics.value.activeWorkflows = result.list?.length || 0
    } catch (error) {
      console.error('Failed to load workflows:', error)
    } finally {
      flowsLoading.value = false
    }
  }

  // 加载所有仪表盘数据
  async function loadDashboardData() {
    loading.value = true

    // 加载工作流（需要先知道数量）
    await loadWorkflows()

    // 并行加载其他数据
    await Promise.all([
      loadStats(),
      loadRecentRuns(),
      loadTriggers(),
    ])

    loading.value = false
  }

  // 是否正在加载
  const isLoading = computed(() =>
    statsLoading.value || runsLoading.value || triggersLoading.value || flowsLoading.value
  )

  return {
    // 状态
    loading,
    isLoading,
    statsLoading,
    runsLoading,
    triggersLoading,
    flowsLoading,

    // 数据
    statistics,
    recentActivity,
    workflowRuns,
    upcomingTriggers,
    activeWorkflows,

    // 方法
    loadDashboardData,
    loadStats,
    loadRecentRuns,
    loadTriggers,
    loadWorkflows,
  }
}
