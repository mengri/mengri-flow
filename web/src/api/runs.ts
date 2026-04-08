import api from './client'
import type {
  Run,
  RunListRequest,
  RunListResponse,
  RunDetail,
  ExecutionTimeline,
  RunStats,
} from '@/types/run'

export const runAPI = {
  // 运行记录列表
  list: (params: RunListRequest) => {
    return api.get<RunListResponse>('/runs', { params })
  },

  // 运行详情
  get: (id: string) => {
    return api.get<RunDetail>(`/runs/${id}`)
  },

  // 执行时间线
  getTimeline: (id: string) => {
    return api.get<ExecutionTimeline>(`/runs/${id}/timeline`)
  },

  // 重试运行
  retry: (id: string) => {
    return api.post<Run>(`/runs/${id}/retry`)
  },

  // 运行统计
  getStats: (params: { workspaceId: string; triggerId?: string }) => {
    return api.get<RunStats>('/runs/stats', { params })
  },
}
