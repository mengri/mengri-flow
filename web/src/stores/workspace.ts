import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/api/client'
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

  // --- Actions ---

  /** 加载工作空间列表 */
  async function loadWorkspaces(): Promise<Workspace[]> {
    loading.value = true
    try {
      const params = { page: 1, pageSize: 100 }
      const result = await api.get<{ total: number; page: number; pageSize: number; list: Workspace[] }>('/workspaces', { params })
      workspaces.value = result.list
      
      // 如果没有当前选中的工作空间，默认选择第一个
      if (!currentWorkspaceId.value && workspaces.value.length > 0) {
        setCurrentWorkspace(workspaces.value[0].id)
      }
      
      return workspaces.value
    } catch (error) {
      console.error('Failed to load workspaces:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  /** 设置当前工作空间 */
  function setCurrentWorkspace(workspaceId: string) {
    currentWorkspaceId.value = workspaceId
    localStorage.setItem('currentWorkspaceId', workspaceId)
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
    // Actions
    loadWorkspaces,
    setCurrentWorkspace,
    getWorkspaceName,
  }
})
