export interface Workspace {
  id: string
  name: string
  description: string
  ownerId: string
  memberCount: number
  createdAt: string
  updatedAt: string
}

export interface CreateWorkspaceRequest {
  name: string
  description?: string
}

export interface UpdateWorkspaceRequest {
  name?: string
  description?: string
}

export interface ListWorkspacesResponse {
  total: number
  page: number
  pageSize: number
  list: Workspace[]
}

export interface WorkspaceMember {
  accountId: string
  email: string
  displayName: string
  role: 'owner' | 'admin' | 'member'
  joinedAt: string
}

export interface AddWorkspaceMemberRequest {
  accountId: string
  role: 'member' | 'admin'
}

export interface ListWorkspaceMembersResponse {
  total: number
  page: number
  pageSize: number
  list: WorkspaceMember[]
}
