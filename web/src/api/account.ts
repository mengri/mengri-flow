import api from './client'
import type {
  CreateAccountRequest,
  AccountResponse,
  AccountDetailResponse,
  ListAccountsRequest,
  ListAccountsResponse,
  ChangeStatusRequest,
  ResendActivationRequest,
  ResendActivationResponse,
  ProfileResponse,
  ChangePasswordRequest,
  ChangePasswordResponse,
  SecurityVerifyRequest,
  SecurityTicketResponse,
  IdentityListResponse,
  AuditEventFilter,
  AuditEventListResponse,
  LoginHistoryItem,
} from '@/types'

// ==================== Admin 账号管理 ====================

/** 创建账号 */
export function createAccount(data: CreateAccountRequest) {
  return api.post<AccountResponse>(
    '/admin/accounts',
    data,
  )
}

/** 账号列表 */
export function listAccounts(params: ListAccountsRequest) {
  return api.get<ListAccountsResponse>(
    '/admin/accounts',
    { params },
  )
}

/** 账号详情 */
export function getAccountDetail(accountId: string) {
  return api.get<AccountDetailResponse>(
    `/admin/accounts/${accountId}`,
  )
}

/** 变更账号状态 */
export function changeAccountStatus(accountId: string, data: ChangeStatusRequest) {
  return api.put<{ accountId: string; status: string; updatedAt: string }>(
    `/admin/accounts/${accountId}/status`,
    data,
  )
}

/** 重发激活邮件 */
export function resendActivation(accountId: string, data?: ResendActivationRequest) {
  return api.post<ResendActivationResponse>(
    `/admin/accounts/${accountId}/activation/resend`,
    data,
  )
}

/** 审计事件列表 */
export function listAuditEvents(params: AuditEventFilter) {
  return api.get<AuditEventListResponse>(
    '/admin/audit/events',
    { params },
  )
}

// ==================== Me 用户中心 ====================

/** 获取我的资料 */
export function getProfile() {
  return api.get<ProfileResponse>('/me/profile')
}

/** 获取我的身份列表 */
export function listMyIdentities() {
  return api.get<IdentityListResponse>('/me/identities')
}

/** 修改密码 */
export function changePassword(data: ChangePasswordRequest) {
  return api.post<ChangePasswordResponse>(
    '/me/password/change',
    data,
  )
}

/** 二次安全验证 */
export function securityVerify(data: SecurityVerifyRequest) {
  return api.post<SecurityTicketResponse>(
    '/me/security/verify',
    data,
  )
}

/** 登录历史 */
export function getLoginHistory() {
  return api.get<AuditEventListResponse>(
    '/me/security/logins',
  )
}
