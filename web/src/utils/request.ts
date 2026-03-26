import axios from 'axios'
import type { ApiResponse } from '@/types'
import { ElMessage } from 'element-plus'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

const TOKEN_KEY = 'accessToken'
const REFRESH_TOKEN_KEY = 'refreshToken'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(accessToken: string, refreshToken: string): void {
  localStorage.setItem(TOKEN_KEY, accessToken)
  localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

export function clearTokens(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
}

// 请求拦截器 — 自动附加 JWT
request.interceptors.request.use(
  (config) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  },
)

// 响应拦截器 — 统一处理 { code, data, msg } 格式
request.interceptors.response.use(
  (response) => {
    const res = response.data as ApiResponse
    if (res.code !== 0) {
      // Session 过期跳转登录
      if (res.code === 110005) {
        clearTokens()
        const current = window.location.pathname
        if (current !== '/login') {
          window.location.href = '/login'
        }
        return Promise.reject(new Error(res.msg || 'Session expired'))
      }
      ElMessage.error(res.msg || 'Request failed')
      return Promise.reject(new Error(res.msg || 'Request failed'))
    }
    return response
  },
  (error) => {
    if (error.response?.status === 401) {
      clearTokens()
      const current = window.location.pathname
      if (current !== '/login') {
        window.location.href = '/login'
      }
      return Promise.reject(error)
    }
    const msg = error.response?.data?.msg || error.message || 'Network error'
    ElMessage.error(msg)
    return Promise.reject(error)
  },
)

export default request
