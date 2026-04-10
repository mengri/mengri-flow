// --- 通用 ---

/** 统一 API 响应信封 */
export interface ApiResponse<T = unknown> {
  code: number
  data: T
  msg: string
}

/** 分页响应 */
export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
}

// --- 账号状态 ---

export type AccountStatus = 'PENDING_ACTIVATION' | 'ACTIVE' | 'LOCKED' | 'DISABLED'

export type AccountRole = 'user' | 'admin'

export type LoginType = 'password' | 'sms' | 'wechat_qr' | 'lark_qr' | 'github_oauth'

// --- 认证 DTO ---

export interface ActivationValidateResponse {
  valid: boolean
  emailMasked: string
  expireAt: string
  alreadyActivated: boolean
}

export interface ActivationConfirmRequest {
  token: string
  password: string
  confirmPassword: string
}

export interface ActivationConfirmResponse {
  activated: boolean
  accountId: string
  status: string
  activatedAt: string
}

export interface DeviceInfo {
  ua: string
  ip: string
  deviceId: string
}

export interface PasswordLoginRequest {
  account: string
  password: string
  deviceInfo?: DeviceInfo
}

export interface LoginResponse {
  accessToken: string
  refreshToken: string
  expiresIn: number
  tokenType: string
  account: AccountBrief
}

export interface AccountBrief {
  accountId: string
  email: string
  displayName: string
  status: AccountStatus
}

export interface RefreshTokenRequest {
  refreshToken: string
}

// --- 账号管理 DTO ---

export interface CreateAccountRequest {
  email: string
  displayName: string
  username: string
}

export interface AccountResponse {
  accountId: string
  email: string
  username: string
  displayName: string
  status: AccountStatus
  role: AccountRole
  activationExpireAt?: string
  activatedAt?: string
  createdAt: string
  updatedAt: string
}

export interface AccountDetailResponse extends AccountResponse {
  identities: IdentityBrief[]
}

export interface ListAccountsRequest {
  page?: number
  pageSize?: number
  status?: AccountStatus | ''
  keyword?: string
}

export interface ListAccountsResponse extends PaginatedResponse<AccountResponse> {}

export interface ChangeStatusRequest {
  action: 'lock' | 'unlock' | 'disable' | 'enable'
  reason?: string
}

export interface ResendActivationRequest {
  reason?: string
}

export interface ResendActivationResponse {
  sent: boolean
  activationExpireAt: string
  throttleSec: number
}

// --- 用户中心 DTO ---

export interface ProfileResponse {
  accountId: string
  email: string
  username: string
  displayName: string
  accountStatus: AccountStatus
  role: AccountRole
}

export interface ChangePasswordRequest {
  oldPassword: string
  newPassword: string
  confirmPassword: string
  revokeOtherSessions?: boolean
}

export interface ChangePasswordResponse {
  changed: boolean
  revokedSessions: number
}

export interface SecurityVerifyRequest {
  password: string
}

export interface SecurityTicketResponse {
  securityTicket: string
  expireAt: string
  ttlSec: number
}

// --- 身份 DTO ---

export interface IdentityBrief {
  identityId: string
  loginType: LoginType
  maskedIdentifier?: string
  boundAt: string
}

export interface IdentityListResponse {
  identities: IdentityBrief[]
  canUnbindLast: boolean
}

// --- 审计 DTO ---

export interface AuditEventItem {
  id: string
  eventType: string
  result: string
  ip: string
  ua: string
  createdAt: string
}

export interface AuditEventFilter {
  accountId?: string
  eventType?: string
  from?: string
  to?: string
  page?: number
  pageSize?: number
}

export interface AuditEventListResponse {
  items: AuditEventItem[]
  total: number
  page: number
  pageSize: number
}

// --- 登录历史 ---

export interface LoginHistoryItem {
  id: string
  eventType: string
  result: string
  ip: string
  ua: string
  createdAt: string
}
