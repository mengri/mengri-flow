export interface Tool {
  id: string
  name: string
  resourceId: string
  type: string
  method: string
  path?: string
  inputSchema: Record<string, any>
  outputSchema: Record<string, any>
  description: string
  tags: string[]
  workspaceId: string
  status: 'draft' | 'published' | 'deprecated'
  currentVersion: number
  createdAt: string
  updatedAt: string
}

export interface CreateToolRequest {
  name: string
  resourceId: string
  type: string
  method: string
  path?: string
  inputSchema: Record<string, any>
  outputSchema: Record<string, any>
  description?: string
  tags?: string[]
  workspaceId: string
}

export interface ImportToolsRequest {
  resourceId: string
  fileType: 'openapi' | 'proto' | 'sqlc'
  fileData: string
}
