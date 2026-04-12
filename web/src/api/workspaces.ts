import api from './client'
import type {
  Workspace,
  CreateWorkspaceRequest,
  UpdateWorkspaceRequest,
  ListWorkspacesResponse,
  WorkspaceMember,
  AddWorkspaceMemberRequest,
  ListWorkspaceMembersResponse,
} from '@/types/workspace'

export const workspaceAPI = {
  /** 工作空间列表 */
  list: (params?: { page?: number; pageSize?: number }) => {
    return api.get<ListWorkspacesResponse>('/workspaces', { params })
  },

  /** 创建工作空间 */
  create: (data: CreateWorkspaceRequest) => {
    return api.post<Workspace>('/workspaces', data)
  },

  /** 获取工作空间详情 */
  get: (id: string) => {
    return api.get<Workspace>(`/workspaces/${id}`)
  },

  /** 更新工作空间 */
  update: (id: string, data: UpdateWorkspaceRequest) => {
    return api.put<Workspace>(`/workspaces/${id}`, data)
  },

  /** 删除工作空间 */
  delete: (id: string) => {
    return api.delete(`/workspaces/${id}`)
  },

  /** 获取工作空间成员列表 */
  listMembers: (id: string, params?: { page?: number; pageSize?: number }) => {
    return api.get<ListWorkspaceMembersResponse>(`/workspaces/${id}/members`, { params })
  },

  /** 添加工作空间成员 */
  addMember: (id: string, data: AddWorkspaceMemberRequest) => {
    return api.post<WorkspaceMember>(`/workspaces/${id}/members`, data)
  },

  /** 移除工作空间成员 */
  removeMember: (id: string, memberId: string) => {
    return api.delete(`/workspaces/${id}/members/${memberId}`)
  },
}
