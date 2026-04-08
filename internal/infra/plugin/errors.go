package plugin

import (
	"errors"
	"fmt"
)

// 预定义的插件错误类型
var (
	ErrConnectionFailed     = errors.New("connection failed")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrTimeout              = errors.New("execution timeout")
	ErrInvalidInput         = errors.New("invalid input")
	ErrNotFound             = errors.New("resource not found")
	ErrRateLimited          = errors.New("rate limited")
	ErrInternal             = errors.New("internal error")
)

// PluginError 插件错误包装器
type PluginError struct {
	Type    string // 错误类型标识
	Message string // 错误描述
	Cause   error  // 原始错误
}

// Error 实现 error 接口
func (e *PluginError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %w", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap 支持错误链
func (e *PluginError) Unwrap() error {
	return e.Cause
}

// NewPluginError 创建新的插件错误
func NewPluginError(errType, message string, cause error) *PluginError {
	return &PluginError{
		Type:    errType,
		Message: message,
		Cause:   cause,
	}
}

// MapErrorToHTTPStatus 将插件错误映射到HTTP状态码
func MapErrorToHTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return 400
	case errors.Is(err, ErrAuthenticationFailed):
		return 401
	case errors.Is(err, ErrNotFound):
		return 404
	case errors.Is(err, ErrRateLimited):
		return 429
	case errors.Is(err, ErrConnectionFailed),
		errors.Is(err, ErrTimeout),
		errors.Is(err, ErrInternal):
		return 500
	default:
		return 500
	}
}
