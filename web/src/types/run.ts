export interface Run {
  id: string
  triggerId: string
  triggerName: string
  flowId: string
  flowName: string
  flowVersion: number
  status: 'success' | 'failed' | 'running' | 'timeout'
  inputData: Record<string, any>
  outputData: Record<string, any>
  errorMessage?: string
  startedAt: string
  finishedAt?: string
  durationMs: number
  clusterId: string
  nodeLogs: NodeLog[]
}

export interface NodeLog {
  nodeId: string
  toolId: string
  toolName: string
  status: 'success' | 'failed' | 'running'
  inputData: Record<string, any>
  outputData: Record<string, any>
  errorMessage?: string
  startedAt: string
  finishedAt?: string
  durationMs: number
}

export interface ExecutionTimeline {
  runId: string
  timeline: TimelineEvent[]
}

export interface TimelineEvent {
  id: string
  event: string
  nodeId?: string
  timestamp: string
  duration?: number
  status?: string
}

export interface RunListRequest {
  workspaceId: string
  triggerId?: string
  status?: string
  page: number
  pageSize: number
  startTime?: string
  endTime?: string
}

export interface RunListResponse {
  list: Run[]
  total: number
  page: number
  pageSize: number
}

export interface RunDetail extends Run {
  executionLogs: ExecutionLog[]
}

export interface ExecutionLog {
  id: string
  level: 'info' | 'warning' | 'error'
  message: string
  timestamp: string
  nodeId?: string
}

export interface RunStats {
  totalRuns: number
  successRate: number
  avgDuration: number
  todayRuns: number
  trend: Array<{
    date: string
    success: number
    failed: number
  }>
}
