/** 统一后端响应格式 */
export interface ApiResponse<T = unknown> {
  code: number
  data: T
  msg: string
}

/** 用户实体 */
export interface User {
  id: number
  username: string
  email: string
  status: number
  created_at: string
  updated_at: string
}

/** 创建用户请求 */
export interface CreateUserRequest {
  username: string
  email: string
  password: string
}

/** 更新用户请求 */
export interface UpdateUserRequest {
  username?: string
  email?: string
}

/** 分页列表响应 */
export interface ListResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}
