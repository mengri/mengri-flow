import api from './client'
import type { Flow, CreateFlowRequest } from '@/types/flow'

export const flowAPI = {
  // 流程列表
  list: (params: { workspaceId: string; status?: string }) => {
    return api.get<Flow[]>('/flows', { params })
  },

  // 创建流程
  create: (data: CreateFlowRequest) => {
    return api.post<Flow>('/flows', data)
  },

  // 获取流程详情
  get: (id: string) => {
    return api.get<Flow>(`/flows/${id}`)
  },

  // 更新流程
  update: (id: string, data: Partial<Flow>) => {
    return api.put<Flow>(`/flows/${id}`, data)
  },

  // 删除流程
  delete: (id: string) => {
    return api.delete(`/flows/${id}`)
  },

  // 测试运行流程
  test: (data: { flowId: string; input: Record<string, any> }) => {
    return api.post('/flows/test', data)
  },

  // 发布流程
  publish: (id: string) => {
    return api.post<Flow>(`/flows/${id}/publish`)
  },
}
