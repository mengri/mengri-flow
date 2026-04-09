import api from './client'
import type {
  ActivationValidateResponse,
  ActivationConfirmRequest,
  ActivationConfirmResponse,
  PasswordLoginRequest,
  LoginResponse,
  RefreshTokenRequest,
} from '@/types'

/** 验证激活链接 */
export function validateActivation(token: string) {
  return api.get<ActivationValidateResponse>(
    '/auth/activation/validate',
    { params: { token } },
  )
}

/** 确认激活并设置密码 */
export function confirmActivation(data: ActivationConfirmRequest) {
  return api.post<ActivationConfirmResponse>(
    '/auth/activation/confirm',
    data,
  )
}

/** 密码登录 */
export function loginByPassword(data: PasswordLoginRequest) {
  return api.post<LoginResponse>(
    '/auth/login/password',
    data,
  )
}

/** 刷新 Token */
export function refreshToken(data: RefreshTokenRequest) {
  return api.post<LoginResponse>(
    '/auth/token/refresh',
    data,
  )
}

/** 登出 */
export function logout() {
  return api.post<{ success: boolean }>(
    '/auth/logout',
  )
}
