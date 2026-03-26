import request from '@/utils/request'
import type {
  ApiResponse,
  ActivationValidateResponse,
  ActivationConfirmRequest,
  ActivationConfirmResponse,
  PasswordLoginRequest,
  LoginResponse,
  RefreshTokenRequest,
} from '@/types'

/** 验证激活链接 */
export function validateActivation(token: string) {
  return request.get<ApiResponse<ActivationValidateResponse>>(
    '/auth/activation/validate',
    { params: { token } },
  )
}

/** 确认激活并设置密码 */
export function confirmActivation(data: ActivationConfirmRequest) {
  return request.post<ApiResponse<ActivationConfirmResponse>>(
    '/auth/activation/confirm',
    data,
  )
}

/** 密码登录 */
export function loginByPassword(data: PasswordLoginRequest) {
  return request.post<ApiResponse<LoginResponse>>(
    '/auth/login/password',
    data,
  )
}

/** 刷新 Token */
export function refreshToken(data: RefreshTokenRequest) {
  return request.post<ApiResponse<LoginResponse>>(
    '/auth/token/refresh',
    data,
  )
}

/** 登出 */
export function logout() {
  return request.post<ApiResponse<{ success: boolean }>>(
    '/auth/logout',
  )
}
