import api from './client'
import type { Trigger, CreateTriggerRequest } from '@/types/trigger'

export const triggerAPI = {
  // 触发器列表
  list: (params: { workspaceId: string }) => {
    return api.get<Trigger[]>('/triggers', { params })
  },

  // 创建触发器
  create: (data: CreateTriggerRequest) => {
    return api.post<Trigger>('/triggers', data)
  },

  // 获取触发器详情
  get: (id: string) => {
    return api.get<Trigger>(`/triggers/${id}`)
  },

  // 更新触发器
  update: (id: string, data: Partial<Trigger>) => {
    return api.put<Trigger>(`/triggers/${id}`, data)
  },

  // 删除触发器
  delete: (id: string) => {
    return api.delete(`/triggers/${id}`)
  },

  // 发布触发器到集群
  publish: (id: string, clusterId: string) => {
    return api.post(`/triggers/${id}/publish`, { clusterId })
  },

  // 停止触发器
  stop: (id: string) => {
    return api.post(`/triggers/${id}/stop`)
  },

  // 启动触发器
  start: (id: string) => {
    return api.post(`/triggers/${id}/start`)
  },
}
