import axios, { AxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'
import type { ApiResponse } from '@/types'

const instance = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

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
  (error) => {
    if (error.response?.status === 401) {
      // Token过期，跳转到登录
      router.push('/login')
    }
    
    ElMessage.error(error.message || '网络错误')
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
