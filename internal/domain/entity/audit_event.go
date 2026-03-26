package entity

import (
	"time"

	"mengri-flow/internal/domain/errors"
)

// 审计事件类型常量
const (
	// 账号管理
	AuditAccountCreated        = "ACCOUNT_CREATED"
	AuditActivationEmailSent   = "ACTIVATION_EMAIL_SENT"
	AuditActivationEmailResent = "ACTIVATION_EMAIL_RESENT"
	AuditAccountActivated      = "ACCOUNT_ACTIVATED"
	AuditAccountLocked         = "ACCOUNT_LOCKED"
	AuditAccountUnlocked       = "ACCOUNT_UNLOCKED"
	AuditAccountDisabled       = "ACCOUNT_DISABLED"
	AuditAccountEnabled        = "ACCOUNT_ENABLED"

	// 认证
	AuditLoginSuccess   = "LOGIN_SUCCESS"
	AuditLoginFailed    = "LOGIN_FAILED"
	AuditLogout         = "LOGOUT"
	AuditTokenRefreshed = "TOKEN_REFRESHED"

	// 身份绑定
	AuditIdentityBound   = "IDENTITY_BOUND"
	AuditIdentityUnbound = "IDENTITY_UNBOUND"

	// 安全操作
	AuditPasswordChanged  = "PASSWORD_CHANGED"
	AuditSecurityVerified = "SECURITY_VERIFIED"
	AuditSessionsRevoked  = "SESSIONS_REVOKED"
)

// 审计事件结果
const (
	AuditResultSuccess = "success"
	AuditResultFailure = "failure"
)

// AuditEvent 审计事件实体。
type AuditEvent struct {
	ID              string
	ActorID         string // 操作人（管理员或用户自己）
	TargetAccountID string // 被操作的账号
	EventType       string
	Result          string // "success" | "failure"
	IP              string
	UA              string
	Metadata        string // JSON 扩展信息
	CreatedAt       time.Time
}

// NewAuditEvent 创建审计事件。
func NewAuditEvent(actorID, targetID, eventType, result, ip, ua string) (*AuditEvent, error) {
	if eventType == "" {
		return nil, errors.ErrNotFound
	}
	return &AuditEvent{
		ActorID:         actorID,
		TargetAccountID: targetID,
		EventType:       eventType,
		Result:          result,
		IP:              ip,
		UA:              ua,
		CreatedAt:       time.Now(),
	}, nil
}
