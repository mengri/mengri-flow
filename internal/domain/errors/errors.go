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
)
