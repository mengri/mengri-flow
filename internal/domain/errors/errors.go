package errors

import "errors"

// 领域层错误定义 — 只使用标准库 errors，不依赖任何外部包。

var (
	// 通用错误
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("resource already exists")

	// User 领域错误
	ErrInvalidUsername = errors.New("invalid username: must be 2-50 characters")
	ErrInvalidEmail    = errors.New("invalid email address")
	ErrInvalidPassword = errors.New("invalid password: must be at least 8 characters")
	ErrUserBanned      = errors.New("user is banned and cannot be activated")
	ErrUserNotFound    = errors.New("user not found")
	ErrEmailTaken      = errors.New("email is already taken")

	// Account 领域错误
	ErrInvalidDisplayName = errors.New("invalid display name: must not be empty")
	ErrAccountNotFound    = errors.New("account not found")
	ErrUsernameTaken      = errors.New("username is already taken")
)

// 通用认证错误
var (
	ErrCredentialsInvalid  = errors.New("credentials invalid")
	ErrAccountNotActivated = errors.New("account not activated")
	ErrAccountLocked       = errors.New("account is locked")
	ErrAccountDisabled     = errors.New("account is disabled")
	ErrSessionExpired      = errors.New("session expired or invalid")
)

// 激活错误
var (
	ErrActivationTokenInvalid = errors.New("activation token invalid")
	ErrActivationTokenExpired = errors.New("activation token expired")
	ErrActivationTokenUsed    = errors.New("activation token already used")
	ErrAlreadyActivated       = errors.New("account already activated")
	ErrActivationTooFrequent  = errors.New("activation email sent too frequently")
)

// 验证码/风控错误
var (
	ErrOTPInvalid      = errors.New("otp code invalid")
	ErrOTPExpired      = errors.New("otp code expired")
	ErrOTPTooFrequent  = errors.New("otp code sent too frequently")
	ErrCaptchaRequired = errors.New("captcha verification required")
	ErrCaptchaFailed   = errors.New("captcha verification failed")
)

// 绑定错误
var (
	ErrIdentityNotBound      = errors.New("identity not bound to any account")
	ErrIdentityAlreadyBound  = errors.New("identity already bound to another account")
	ErrPhoneAlreadyBound     = errors.New("phone already bound to another account")
	ErrCannotUnbindLast      = errors.New("cannot unbind the last login method")
	ErrBindTicketInvalid     = errors.New("bind ticket invalid or expired")
	ErrSecurityTicketInvalid = errors.New("security ticket invalid or expired")
)

// 状态迁移错误
var (
	ErrInvalidStatusTransition = errors.New("invalid account status transition")
)
