import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { workspaceAPI } from '@/api/workspaces'
import type { Workspace } from '@/types/workspace'

export const useWorkspaceStore = defineStore('workspace', () => {
  // --- State ---
  const workspaces = ref<Workspace[]>([])
  const loading = ref(false)
  const currentWorkspaceId = ref<string | null>(localStorage.getItem('currentWorkspaceId'))

  // --- Getters ---
  const currentWorkspace = computed(() =>
    workspaces.value.find(ws => ws.id === currentWorkspaceId.value) || null
  )

  const currentWorkspaceIdOrThrow = computed(() => {
    if (!currentWorkspaceId.value) {
      throw new Error('No workspace selected')
    }
    return currentWorkspaceId.value
  })

  /** 是否已有 localStorage 记录的 workspace（但不一定在列表中有效） */
  const hasStoredWorkspace = computed(() => !!currentWorkspaceId.value)

  /** 是否有当前选中的有效 workspace */
  const hasCurrentWorkspace = computed(() => !!currentWorkspace.value)

  // --- Actions ---

  /**
   * 加载工作空间列表
   *
   * 行为：
   * - localStorage 中有记录且列表中仍存在 → 保持选中
   * - localStorage 中有记录但列表中已不存在 → 清除记录（不自动选择）
   * - localStorage 中无记录（首次登录/已清除） → 不自动选择，等待用户手动选择
   *
   * 返回 'selected' | 'none' 表示是否已有选中
   */
  async function loadWorkspaces(): Promise<'selected' | 'none'> {
    loading.value = true
    try {
      const result = await workspaceAPI.list({ page: 1, pageSize: 100 })
      workspaces.value = result.list

      if (currentWorkspaceId.value && !workspaces.value.find(ws => ws.id === currentWorkspaceId.value)) {
        // 之前选中的 workspace 已不存在，清除记录
        currentWorkspaceId.value = null
        localStorage.removeItem('currentWorkspaceId')
      }

      return hasCurrentWorkspace.value ? 'selected' : 'none'
    } catch (error) {
      console.error('Failed to load workspaces:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  /** 创建工作空间 */
  async function createWorkspace(data: { name: string; description?: string }): Promise<Workspace> {
    const workspace = await workspaceAPI.create(data)
    workspaces.value.unshift(workspace)
    return workspace
  }

  /** 更新工作空间 */
  async function updateWorkspace(id: string, data: { name?: string; description?: string }): Promise<Workspace> {
    const updated = await workspaceAPI.update(id, data)
    const index = workspaces.value.findIndex(ws => ws.id === id)
    if (index !== -1) {
      workspaces.value[index] = updated
    }
    return updated
  }

  /** 删除工作空间 */
  async function deleteWorkspace(id: string): Promise<void> {
    await workspaceAPI.delete(id)
    workspaces.value = workspaces.value.filter(ws => ws.id !== id)

    // 如果删除的是当前选中的工作空间，切换到第一个
    if (currentWorkspaceId.value === id) {
      if (workspaces.value.length > 0) {
        setCurrentWorkspace(workspaces.value[0].id)
      } else {
        currentWorkspaceId.value = null
        localStorage.removeItem('currentWorkspaceId')
      }
    }
  }

  /** 设置当前工作空间 */
  function setCurrentWorkspace(workspaceId: string) {
    currentWorkspaceId.value = workspaceId
    localStorage.setItem('currentWorkspaceId', workspaceId)
  }

  /** 清除当前工作空间选择（登出时调用） */
  function clearCurrentWorkspace() {
    currentWorkspaceId.value = null
    localStorage.removeItem('currentWorkspaceId')
    workspaces.value = []
  }

  /** 获取工作空间名称 */
  function getWorkspaceName(workspaceId: string): string {
    const workspace = workspaces.value.find(ws => ws.id === workspaceId)
    return workspace?.name || workspaceId
  }

  return {
    // State
    workspaces,
    loading,
    currentWorkspaceId,
    // Getters
    currentWorkspace,
    currentWorkspaceIdOrThrow,
    hasStoredWorkspace,
    hasCurrentWorkspace,
    // Actions
    loadWorkspaces,
    createWorkspace,
    updateWorkspace,
    deleteWorkspace,
    setCurrentWorkspace,
    clearCurrentWorkspace,
    getWorkspaceName,
  }
})
