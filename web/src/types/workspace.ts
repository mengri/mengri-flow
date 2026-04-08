export interface Workspace {
  id: string
  name: string
  description?: string
  status: 'active' | 'inactive'
  createdAt: string
  updatedAt: string
}

export interface CreateWorkspaceRequest {
  name: string
  description?: string
}
