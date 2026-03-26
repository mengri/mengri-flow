import request from '@/utils/request'
import type {
  ApiResponse,
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
  return request.post<ApiResponse<AccountResponse>>(
    '/admin/accounts',
    data,
  )
}

/** 账号列表 */
export function listAccounts(params: ListAccountsRequest) {
  return request.get<ApiResponse<ListAccountsResponse>>(
    '/admin/accounts',
    { params },
  )
}

/** 账号详情 */
export function getAccountDetail(accountId: string) {
  return request.get<ApiResponse<AccountDetailResponse>>(
    `/admin/accounts/${accountId}`,
  )
}

/** 变更账号状态 */
export function changeAccountStatus(accountId: string, data: ChangeStatusRequest) {
  return request.put<ApiResponse<{ accountId: string; status: string; updatedAt: string }>>(
    `/admin/accounts/${accountId}/status`,
    data,
  )
}

/** 重发激活邮件 */
export function resendActivation(accountId: string, data?: ResendActivationRequest) {
  return request.post<ApiResponse<ResendActivationResponse>>(
    `/admin/accounts/${accountId}/activation/resend`,
    data,
  )
}

/** 审计事件列表 */
export function listAuditEvents(params: AuditEventFilter) {
  return request.get<ApiResponse<AuditEventListResponse>>(
    '/admin/audit/events',
    { params },
  )
}

// ==================== Me 用户中心 ====================

/** 获取我的资料 */
export function getProfile() {
  return request.get<ApiResponse<ProfileResponse>>('/me/profile')
}

/** 获取我的身份列表 */
export function listMyIdentities() {
  return request.get<ApiResponse<IdentityListResponse>>('/me/identities')
}

/** 修改密码 */
export function changePassword(data: ChangePasswordRequest) {
  return request.post<ApiResponse<ChangePasswordResponse>>(
    '/me/password/change',
    data,
  )
}

/** 二次安全验证 */
export function securityVerify(data: SecurityVerifyRequest) {
  return request.post<ApiResponse<SecurityTicketResponse>>(
    '/me/security/verify',
    data,
  )
}

/** 登录历史 */
export function getLoginHistory() {
  return request.get<ApiResponse<LoginHistoryItem[]>>(
    '/me/security/logins',
  )
}
