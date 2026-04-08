import api from './client'
import type { Tool, CreateToolRequest, ImportToolsRequest } from '@/types/tool'

export const toolAPI = {
  // 工具列表
  list: (params: { workspaceId: string; resourceId?: string; status?: string }) => {
    return api.get<Tool[]>('/tools', { params })
  },

  // 创建工具
  create: (data: CreateToolRequest) => {
    return api.post<Tool>('/tools', data)
  },

  // 获取工具详情
  get: (id: string) => {
    return api.get<Tool>(`/tools/${id}`)
  },

  // 更新工具
  update: (id: string, data: Partial<CreateToolRequest>) => {
    return api.put<Tool>(`/tools/${id}`, data)
  },

  // 删除工具
  delete: (id: string) => {
    return api.delete(`/tools/${id}`)
  },

  // 测试工具
  test: (data: { toolId: string; input: Record<string, any> }) => {
    return api.post(`/tools/test`, data)
  },

  // 批量导入工具
  import: (data: FormData) => {
    return api.post<Tool[]>('/tools/import', data, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },

  // 发布工具
  publish: (id: string) => {
    return api.post<Tool>(`/tools/${id}/publish`)
  },

  // 下线工具
  deprecate: (id: string) => {
    return api.post(`/tools/${id}/deprecate`)
  },
}
