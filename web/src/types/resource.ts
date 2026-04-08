export interface Resource {
  id: string
  name: string
  type: 'http' | 'grpc' | 'mysql' | 'postgres'
  config: Record<string, any>
  status: 'active' | 'inactive' | 'error'
  workspaceId: string
  description: string
  createdAt: string
  updatedAt: string
}

export interface CreateResourceRequest {
  name: string
  type: string
  config: Record<string, any>
  workspaceId: string
  description?: string
}
