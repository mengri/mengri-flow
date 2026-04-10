export interface Flow {
  id: string
  name: string
  description: string
  canvasData: {
    nodes: Node[]
    edges: Edge[]
  }
  inputSchema: Record<string, any>
  outputSchema: Record<string, any>
  workspaceId: string
  status: 'draft' | 'active' | 'inactive'
  currentVersion: number
  createdAt: string
  updatedAt: string
}

export interface Node {
  id: string
  type: 'start' | 'end' | 'tool' | 'condition'
  position: { x: number; y: number }
  data: {
    toolId?: string
    toolVersion?: number
    inputMapping?: Array<{ source: string; target: string }>
    config?: Record<string, any>
  }
}

export interface Edge {
  id: string
  source: string
  target: string
  type: string
}

export interface CreateFlowRequest {
  name: string
  description?: string
  workspaceId: string
}
