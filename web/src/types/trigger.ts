export interface Trigger {
  id: string
  name: string
  type: 'restful' | 'timer' | 'rabbitmq' | 'kafka'
  config: Record<string, any>
  flowId: string
  flowVersion: number
  clusterId: string
  inputMapping: Record<string, string>
  outputMapping: Record<string, string>
  errorHandling: {
    strategy: 'default' | 'custom'
    customErrorFormat?: Record<string, any>
    retryOnFailure: boolean
  }
  workspaceId: string
  status: 'active' | 'inactive'
  createdAt: string
  updatedAt: string
}

export interface CreateTriggerRequest {
  name: string
  type: 'restful' | 'timer' | 'rabbitmq' | 'kafka'
  config: Record<string, any>
  flowId: string
  flowVersion: number
  clusterId: string
  inputMapping: Record<string, string>
  outputMapping: Record<string, string>
  errorHandling: {
    strategy: 'default' | 'custom'
    customErrorFormat?: Record<string, any>
    retryOnFailure: boolean
  }
  workspaceId: string
}
