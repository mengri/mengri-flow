import api from './client'
import type { Resource, CreateResourceRequest } from '@/types/resource'
import type { Tool } from '@/types/tool'

export const resourceAPI = {
  // 资源列表
  list: (params: { workspaceId: string; type?: string; status?: string }) => {
    return api.get<Resource[]>('/resources', { params })
  },

  // 创建资源
  create: (data: CreateResourceRequest) => {
    return api.post<Resource>('/resources', data)
  },

  // 获取资源详情
  get: (id: string) => {
    return api.get<Resource>(`/resources/${id}`)
  },

  // 更新资源
  update: (id: string, data: Partial<CreateResourceRequest>) => {
    return api.put<Resource>(`/resources/${id}`, data)
  },

  // 删除资源
  delete: (id: string) => {
    return api.delete(`/resources/${id}`)
  },

  // 测试连接
  testConnection: (data: { type: string; config: Record<string, any> }) => {
    return api.post('/resources/test-connection', data)
  },

  // 提取工具
  extractTools: (id: string) => {
    return api.post<Tool[]>(`/resources/${id}/extract-tools`)
  },
}
