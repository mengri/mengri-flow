package handler

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	domainerrors "mengri-flow/internal/domain/errors"
	"mengri-flow/pkg/response"
)

// handleDomainError 将领域错误映射到HTTP状态码
func handleDomainError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// 记录错误日志
	slog.Error("domain error occurred",
		"module", "http.handler",
		"action", "handle_domain_error",
		"error", err.Error(),
		"success", false,
	)

	// 将领域错误映射到HTTP状态码
	switch {
	// 400 Bad Request
	case errors.Is(err, domainerrors.ErrValidationFailed),
		errors.Is(err, domainerrors.ErrInvalidInput),
		errors.Is(err, domainerrors.ErrInvalidUsername),
		errors.Is(err, domainerrors.ErrInvalidEmail),
		errors.Is(err, domainerrors.ErrInvalidPassword),
		errors.Is(err, domainerrors.ErrInvalidDisplayName),
		errors.Is(err, domainerrors.ErrActivationTokenInvalid),
		errors.Is(err, domainerrors.ErrOTPInvalid),
		errors.Is(err, domainerrors.ErrCaptchaFailed),
		errors.Is(err, domainerrors.ErrBindTicketInvalid),
		errors.Is(err, domainerrors.ErrSecurityTicketInvalid):
		response.BadRequest(c, err.Error())

	// 401 Unauthorized
	case errors.Is(err, domainerrors.ErrUnauthorized),
		errors.Is(err, domainerrors.ErrCredentialsInvalid),
		errors.Is(err, domainerrors.ErrAccountNotActivated),
		errors.Is(err, domainerrors.ErrAccountLocked),
		errors.Is(err, domainerrors.ErrAccountDisabled),
		errors.Is(err, domainerrors.ErrSessionExpired),
		errors.Is(err, domainerrors.ErrActivationTokenExpired),
		errors.Is(err, domainerrors.ErrOTPExpired),
		errors.Is(err, domainerrors.ErrActivationTokenUsed):
		response.Unauthorized(c, err.Error())

	// 403 Forbidden
	case errors.Is(err, domainerrors.ErrForbidden),
		errors.Is(err, domainerrors.ErrInvalidStatusTransition),
		errors.Is(err, domainerrors.ErrIdentityNotBound),
		errors.Is(err, domainerrors.ErrIdentityAlreadyBound),
		errors.Is(err, domainerrors.ErrPhoneAlreadyBound),
		errors.Is(err, domainerrors.ErrCannotUnbindLast):
		response.Forbidden(c, err.Error())

	// 404 Not Found
	case errors.Is(err, domainerrors.ErrNotFound),
		errors.Is(err, domainerrors.ErrAccountNotFound):
		response.NotFound(c, err.Error())

	// 409 Conflict
	case errors.Is(err, domainerrors.ErrConflict),
		errors.Is(err, domainerrors.ErrEmailTaken),
		errors.Is(err, domainerrors.ErrUsernameTaken),
		errors.Is(err, domainerrors.ErrAlreadyActivated),
		errors.Is(err, domainerrors.ErrActivationTooFrequent),
		errors.Is(err, domainerrors.ErrOTPTooFrequent),
		errors.Is(err, domainerrors.ErrCaptchaRequired):
		response.Conflict(c, err.Error())

	// 500 Internal Server Error (默认)
	default:
		response.InternalError(c, fmt.Sprintf("internal server error: %v", err))
	}
}

// handleError 包装错误处理，支持错误链
func handleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// 如果是领域错误，使用领域错误处理
	if isDomainError(err) {
		handleDomainError(c, err)
		return
	}

	// 记录内部错误
	slog.Error("internal error occurred",
		"module", "http.handler",
		"action", "handle_error",
		"error", err.Error(),
		"success", false,
	)

	// 默认返回500错误
	response.InternalError(c, "internal server error")
}

// isDomainError 检查错误是否是领域错误
func isDomainError(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否是预定义的领域错误
	switch {
	case errors.Is(err, domainerrors.ErrNotFound),
		errors.Is(err, domainerrors.ErrUnauthorized),
		errors.Is(err, domainerrors.ErrForbidden),
		errors.Is(err, domainerrors.ErrConflict),
		errors.Is(err, domainerrors.ErrInvalidUsername),
		errors.Is(err, domainerrors.ErrInvalidEmail),
		errors.Is(err, domainerrors.ErrInvalidPassword),
		errors.Is(err, domainerrors.ErrEmailTaken),
		errors.Is(err, domainerrors.ErrAccountNotFound),
		errors.Is(err, domainerrors.ErrUsernameTaken),
		errors.Is(err, domainerrors.ErrCredentialsInvalid),
		errors.Is(err, domainerrors.ErrAccountNotActivated),
		errors.Is(err, domainerrors.ErrAccountLocked),
		errors.Is(err, domainerrors.ErrAccountDisabled),
		errors.Is(err, domainerrors.ErrSessionExpired),
		errors.Is(err, domainerrors.ErrInvalidDisplayName),
		errors.Is(err, domainerrors.ErrActivationTokenInvalid),
		errors.Is(err, domainerrors.ErrActivationTokenExpired),
		errors.Is(err, domainerrors.ErrActivationTokenUsed),
		errors.Is(err, domainerrors.ErrAlreadyActivated),
		errors.Is(err, domainerrors.ErrActivationTooFrequent),
		errors.Is(err, domainerrors.ErrOTPInvalid),
		errors.Is(err, domainerrors.ErrOTPExpired),
		errors.Is(err, domainerrors.ErrOTPTooFrequent),
		errors.Is(err, domainerrors.ErrCaptchaRequired),
		errors.Is(err, domainerrors.ErrCaptchaFailed),
		errors.Is(err, domainerrors.ErrIdentityNotBound),
		errors.Is(err, domainerrors.ErrIdentityAlreadyBound),
		errors.Is(err, domainerrors.ErrPhoneAlreadyBound),
		errors.Is(err, domainerrors.ErrCannotUnbindLast),
		errors.Is(err, domainerrors.ErrInvalidStatusTransition),
		errors.Is(err, domainerrors.ErrBindTicketInvalid),
		errors.Is(err, domainerrors.ErrSecurityTicketInvalid):
		return true
	default:
		return false
	}
}

// DomainErrorWrapper 包装错误并添加上下文
func DomainErrorWrapper(action string) func(error) error {
	return func(err error) error {
		if err == nil {
			return nil
		}
		return fmt.Errorf("%s: %w", action, err)
	}
}
