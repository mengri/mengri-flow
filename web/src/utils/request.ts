export function buildURL(path: string, params: Record<string, any> = {}) {
  const url = new URL(path, window.location.origin)
  
  Object.keys(params).forEach(key => {
    if (params[key] !== undefined && params[key] !== null) {
      url.searchParams.append(key, params[key])
    }
  })
  
  return url.toString()
}

export function downloadFile(content: Blob, filename: string) {
  const url = window.URL.createObjectURL(content)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

import dayjs from 'dayjs'

export function formatDate(date: string | Date, format = 'YYYY-MM-DD HH:mm:ss') {
  return dayjs(date).format(format)
}

export function getToken(): string | null {
  return localStorage.getItem('token')
}

export function setToken(accessToken: string, refreshToken?: string): void {
  localStorage.setItem('token', accessToken)
  if (refreshToken) {
    localStorage.setItem('refreshToken', refreshToken)
  }
}

export function getRefreshToken(): string | null {
  return localStorage.getItem('refreshToken')
}

export function clearTokens(): void {
  localStorage.removeItem('token')
  localStorage.removeItem('refreshToken')
  localStorage.removeItem('currentWorkspaceId')
}

export function formatDuration(ms: number) {
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
  return `${(ms / 60000).toFixed(1)}min`
}

export function statusTagType(status: string) {
  const map: Record<string, string> = {
    active: 'success',
    success: 'success',
    published: 'success',
    error: 'danger',
    failed: 'danger',
    inactive: 'info',
    draft: 'info',
    running: 'warning',
    timeout: 'info',
  }
  return map[status] || 'info'
}

export function statusText(status: string) {
  const map: Record<string, string> = {
    active: '正常',
    success: '成功',
    published: '已发布',
    error: '异常',
    failed: '失败',
    inactive: '未激活',
    draft: '草稿',
    running: '运行中',
    timeout: '超时',
  }
  return map[status] || status
}

// Helper function to safely access objects with string keys
export function getValueByKey<T extends Record<string, any>>(obj: T, key: string): any {
  return obj[key]
}

// Type guard for checking if a property exists on an object
export function hasKey<T extends object>(obj: T, key: string | number | symbol): key is keyof T {
  return key in obj
}

// Safe getter for object properties with string keys
export function safeGet<T extends Record<string, any>, K extends keyof T>(obj: T, key: string, defaultValue?: T[K]): T[K] | undefined {
  if (key in obj) {
    return obj[key as K]
  }
  return defaultValue
}

// Safe getter for status maps with fallbacks
export function getStatusValue(map: Record<string, string>, key: string, fallback = 'info'): string {
  return map[key] || fallback
}
