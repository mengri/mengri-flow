package dto

// CreateAccountRequest 创建账号请求（管理员）
type CreateAccountRequest struct {
	Email       string `json:"email" binding:"required,email"`
	DisplayName string `json:"displayName" binding:"required,min=1,max=50"`
	Username    string `json:"username" binding:"required,min=2,max=50"`
}

// AccountResponse 账号响应
type AccountResponse struct {
	AccountID          string `json:"accountId"`
	Email              string `json:"email"`
	Username           string `json:"username"`
	DisplayName        string `json:"displayName"`
	Status             string `json:"status"`
	Role               string `json:"role"`
	ActivationExpireAt string `json:"activationExpireAt,omitempty"`
	ActivatedAt        string `json:"activatedAt,omitempty"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
}

// AccountDetailResponse 账号详情响应（含身份列表）
type AccountDetailResponse struct {
	AccountResponse
	Identities []IdentityBrief `json:"identities"`
}

// ListAccountsRequest 账号列表请求
type ListAccountsRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=PENDING_ACTIVATION ACTIVE LOCKED DISABLED"`
	Keyword  string `form:"keyword"`
}

// ListAccountsResponse 账号列表响应
type ListAccountsResponse struct {
	Items    []AccountResponse `json:"items"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

// ChangeStatusRequest 管理员变更账号状态请求
type ChangeStatusRequest struct {
	Action string `json:"action" binding:"required,oneof=lock unlock disable enable"`
	Reason string `json:"reason"`
}

// ResendActivationRequest 重发激活邮件请求
type ResendActivationRequest struct {
	Reason string `json:"reason"`
}

// ResendActivationResponse 重发激活邮件响应
type ResendActivationResponse struct {
	Sent               bool   `json:"sent"`
	ActivationExpireAt string `json:"activationExpireAt"`
	ThrottleSec        int    `json:"throttleSec"`
}

// ProfileResponse 用户资料响应
type ProfileResponse struct {
	AccountID     string `json:"accountId"`
	Email         string `json:"email"`
	Username      string `json:"username"`
	DisplayName   string `json:"displayName"`
	AccountStatus string `json:"accountStatus"`
	Role          string `json:"role"`
}

// SecurityTicketResponse 二次验证 ticket 响应
type SecurityTicketResponse struct {
	SecurityTicket string `json:"securityTicket"`
	ExpireAt       string `json:"expireAt"`
	TTLSec         int    `json:"ttlSec"`
}

// SecurityVerifyRequest 二次验证请求
type SecurityVerifyRequest struct {
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword         string `json:"oldPassword" binding:"required"`
	NewPassword         string `json:"newPassword" binding:"required,min=8"`
	ConfirmPassword     string `json:"confirmPassword" binding:"required,eqfield=NewPassword"`
	RevokeOtherSessions bool   `json:"revokeOtherSessions"`
}

// ChangePasswordResponse 修改密码响应
type ChangePasswordResponse struct {
	Changed         bool `json:"changed"`
	RevokedSessions int  `json:"revokedSessions"`
}

// AuditEventItem 审计事件条目
type AuditEventItem struct {
	ID        string `json:"id"`
	EventType string `json:"eventType"`
	Result    string `json:"result"`
	IP        string `json:"ip"`
	UA        string `json:"ua"`
	CreatedAt string `json:"createdAt"`
}

// AuditEventFilter 审计事件查询过滤器
type AuditEventFilter struct {
	AccountID string `form:"accountId"`
	EventType string `form:"eventType"`
	From      string `form:"from"`
	To        string `form:"to"`
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

// AuditEventListResponse 审计事件列表响应
type AuditEventListResponse struct {
	Items    []AuditEventItem `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}
