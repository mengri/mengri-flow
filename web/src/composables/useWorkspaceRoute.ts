import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useWorkspaceStore } from '@/stores/workspace'

/**
 * 工作空间路由辅助 composable
 *
 * 所有空间内的路由格式为 /workspace/:workspaceId/xxx
 * 本 composable 提供统一的路径生成函数，确保所有导航引用都包含正确的 workspaceId
 */
export function useWorkspaceRoute() {
  const route = useRoute()
  const workspaceStore = useWorkspaceStore()

  /** 当前 URL 中的 workspaceId（来自路由参数） */
  const workspaceId = computed(() =>
    (route.params.workspaceId as string) || workspaceStore.currentWorkspaceId || ''
  )

  /**
   * 生成带 workspaceId 前缀的路径
   * @param path - 相对路径，如 '/flows' 或 '/flows/new'
   * @param wsId - 可选指定 workspaceId，默认使用当前路由中的
   */
  function wsPath(path: string, wsId?: string): string {
    const wid = wsId || workspaceId.value
    if (!wid) return path
    // 确保路径以 / 开头
    const normalized = path.startsWith('/') ? path : `/${path}`
    return `/workspace/${wid}${normalized}`
  }

  /**
   * 生成带 workspaceId 前缀的对象路由（用于 router.push/replace）
   */
  function wsRoute(path: string, query?: Record<string, any>, wsId?: string) {
    const result: any = { path: wsPath(path, wsId) }
    if (query) result.query = query
    return result
  }

  // --- 常用路径快捷方法 ---

  function dashboardPath() { return wsPath('/') }
  function flowsPath() { return wsPath('/flows') }
  function flowDetailPath(id: string) { return wsPath(`/flows/${id}`) }
  function createFlowPath() { return wsPath('/flows/new') }

  function toolsPath() { return wsPath('/tools') }
  function toolDetailPath(id: string) { return wsPath(`/tools/${id}`) }
  function createToolPath() { return wsPath('/tools/new') }
  function importToolPath() { return wsPath('/tools/import') }

  function resourcesPath() { return wsPath('/resources') }
  function resourceDetailPath(id: string) { return wsPath(`/resources/${id}`) }
  function createResourcePath() { return wsPath('/resources/new') }

  function triggersPath() { return wsPath('/triggers') }
  function triggerDetailPath(id: string) { return wsPath(`/triggers/${id}`) }
  function createTriggerPath() { return wsPath('/triggers/new') }

  function runsPath() { return wsPath('/runs') }
  function runDetailPath(id: string) { return wsPath(`/runs/${id}`) }

  function settingsPath() { return wsPath('/settings') }
  function settingsMembersPath() { return wsPath('/settings/members') }

  return {
    workspaceId,
    wsPath,
    wsRoute,
    // 快捷路径
    dashboardPath,
    flowsPath,
    flowDetailPath,
    createFlowPath,
    toolsPath,
    toolDetailPath,
    createToolPath,
    importToolPath,
    resourcesPath,
    resourceDetailPath,
    createResourcePath,
    triggersPath,
    triggerDetailPath,
    createTriggerPath,
    runsPath,
    runDetailPath,
    settingsPath,
    settingsMembersPath,
  }
}
