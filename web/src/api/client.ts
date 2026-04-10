import axios, { AxiosRequestConfig, InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'
import type { ApiResponse } from '@/types'
import { clearTokens, getRefreshToken, setToken } from '@/utils/request'
import { refreshToken } from '@/api/auth'

const instance = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 是否正在刷新 token
let isRefreshing = false
// 刷新期间排队的请求
let pendingRequests: Array<{ resolve: (token: string) => void; reject: (err: unknown) => void }> = []

/** 处理刷新队列：成功时重放所有排队请求 */
function processQueue(error: unknown, token: string | null = null) {
  pendingRequests.forEach(({ resolve, reject }) => {
    if (error) {
      reject(error)
    } else {
      resolve(token!)
    }
  })
  pendingRequests = []
}

/** 清理认证状态并跳转登录页 */
function handleAuthExpired() {
  clearTokens()
  // 避免在登录页重复跳转
  if (router.currentRoute.value.path !== '/login') {
    router.push({ path: '/login', query: { redirect: router.currentRoute.value.fullPath } })
  }
}

// 请求拦截器
instance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器
instance.interceptors.response.use(
  (response) => {
    const { code, data, msg } = response.data

    if (code !== 0) {
      ElMessage.error(msg || '请求失败')
      return Promise.reject(new Error(msg))
    }

    return data
  },
  async (error) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    // 401：尝试用 refreshToken 刷新
    if (error.response?.status === 401 && !originalRequest._retry) {
      // 如果是刷新 token 的请求本身失败，直接清理并跳转
      if (originalRequest.url === '/auth/token/refresh') {
        handleAuthExpired()
        return Promise.reject(error)
      }

      if (isRefreshing) {
        // 已经在刷新中，排队等待
        return new Promise((resolve, reject) => {
          pendingRequests.push({
            resolve: (token: string) => {
              originalRequest.headers.Authorization = `Bearer ${token}`
              resolve(instance(originalRequest))
            },
            reject,
          })
        })
      }

      originalRequest._retry = true
      const refreshTokenValue = getRefreshToken()

      if (!refreshTokenValue) {
        // 没有 refreshToken，直接清理跳转
        handleAuthExpired()
        return Promise.reject(error)
      }

      isRefreshing = true
      try {
        const result = await refreshToken({ refreshToken: refreshTokenValue })
        const newAccessToken = (result as any).accessToken
        const newRefreshToken = (result as any).refreshToken
        setToken(newAccessToken, newRefreshToken)

        // 重放排队的请求
        processQueue(null, newAccessToken)

        // 重试原始请求
        originalRequest.headers.Authorization = `Bearer ${newAccessToken}`
        return instance(originalRequest)
      } catch (refreshError) {
        // 刷新失败，清理并跳转
        processQueue(refreshError)
        handleAuthExpired()
        return Promise.reject(refreshError)
      } finally {
        isRefreshing = false
      }
    }

    // 非 401 错误
    if (error.response?.status === 401) {
      // 已经重试过仍然 401
      handleAuthExpired()
    } else {
      ElMessage.error(error.response?.data?.msg || error.message || '网络错误')
    }

    return Promise.reject(error)
  }
)

// 自定义 API 方法，返回解包后的数据
const api = {
  get: <T = any>(url: string, config?: AxiosRequestConfig) => {
    return instance.get<ApiResponse<T>>(url, config).then(res => res as unknown as T)
  },
  post: <T = any>(url: string, data?: any, config?: AxiosRequestConfig) => {
    return instance.post<ApiResponse<T>>(url, data, config).then(res => res as unknown as T)
  },
  put: <T = any>(url: string, data?: any, config?: AxiosRequestConfig) => {
    return instance.put<ApiResponse<T>>(url, data, config).then(res => res as unknown as T)
  },
  delete: <T = any>(url: string, config?: AxiosRequestConfig) => {
    return instance.delete<ApiResponse<T>>(url, config).then(res => res as unknown as T)
  },
}

export default api
