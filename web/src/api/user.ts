import request from '@/utils/request'
import type { ApiResponse, User, CreateUserRequest, UpdateUserRequest, ListResponse } from '@/types'

/** 获取用户列表 */
export function getUserList(page = 1, pageSize = 20) {
  return request.get<ApiResponse<ListResponse<User>>>('/users', {
    params: { page, page_size: pageSize },
  })
}

/** 获取用户详情 */
export function getUserById(id: number) {
  return request.get<ApiResponse<User>>(`/users/${id}`)
}

/** 创建用户 */
export function createUser(data: CreateUserRequest) {
  return request.post<ApiResponse<User>>('/users', data)
}

/** 更新用户 */
export function updateUser(id: number, data: UpdateUserRequest) {
  return request.put<ApiResponse<User>>(`/users/${id}`, data)
}

/** 删除用户 */
export function deleteUser(id: number) {
  return request.delete<ApiResponse<null>>(`/users/${id}`)
}
