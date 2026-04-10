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
