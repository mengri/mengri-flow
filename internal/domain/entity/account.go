package entity

import (
	"mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/valueobject"
	"time"
)

// AccountStatus 账号状态枚举
type AccountStatus string

const (
	AccountStatusPendingActivation AccountStatus = "PENDING_ACTIVATION"
	AccountStatusActive            AccountStatus = "ACTIVE"
	AccountStatusLocked            AccountStatus = "LOCKED"
	AccountStatusDisabled          AccountStatus = "DISABLED"
)

// Account 是账号聚合根，包含状态机与业务方法。
type Account struct {
	ID           string
	Email        valueobject.Email
	Username     string
	DisplayName  string
	Status       AccountStatus
	PasswordHash string // 已哈希密码，PENDING 时为空
	Role         string // "user" | "admin"
	ActivatedAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewAccount 创建未激活账号（管理员调用）。
func NewAccount(email, username, displayName string) (*Account, error) {
	if username == "" || len(username) < 2 || len(username) > 50 {
		return nil, errors.ErrInvalidUsername
	}
	if displayName == "" {
		return nil, errors.ErrInvalidDisplayName
	}

	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &Account{
		Email:       emailVO,
		Username:    username,
		DisplayName: displayName,
		Status:      AccountStatusPendingActivation,
		Role:        "user",
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Activate 激活账号并设置密码。
// 前置：Status == PENDING_ACTIVATION
// 后置：Status → ACTIVE, PasswordHash 设置, ActivatedAt 设置
func (a *Account) Activate(hashedPassword string) error {
	if a.Status != AccountStatusPendingActivation {
		return errors.ErrAlreadyActivated
	}
	if hashedPassword == "" {
		return errors.ErrInvalidPassword
	}
	now := time.Now()
	a.Status = AccountStatusActive
	a.PasswordHash = hashedPassword
	a.ActivatedAt = &now
	a.UpdatedAt = now
	return nil
}

// Lock 锁定账号。
// 前置：Status == ACTIVE
func (a *Account) Lock() error {
	if a.Status != AccountStatusActive {
		return errors.ErrInvalidStatusTransition
	}
	a.Status = AccountStatusLocked
	a.UpdatedAt = time.Now()
	return nil
}

// Unlock 解锁账号。
// 前置：Status == LOCKED
func (a *Account) Unlock() error {
	if a.Status != AccountStatusLocked {
		return errors.ErrInvalidStatusTransition
	}
	a.Status = AccountStatusActive
	a.UpdatedAt = time.Now()
	return nil
}

// Disable 禁用账号。
// 前置：Status == ACTIVE || Status == LOCKED
func (a *Account) Disable() error {
	if a.Status != AccountStatusActive && a.Status != AccountStatusLocked {
		return errors.ErrInvalidStatusTransition
	}
	a.Status = AccountStatusDisabled
	a.UpdatedAt = time.Now()
	return nil
}

// Enable 恢复账号。
// 前置：Status == DISABLED
func (a *Account) Enable() error {
	if a.Status != AccountStatusDisabled {
		return errors.ErrInvalidStatusTransition
	}
	a.Status = AccountStatusActive
	a.UpdatedAt = time.Now()
	return nil
}

// CanLogin 是否允许登录（仅 ACTIVE 状态）。
func (a *Account) CanLogin() bool {
	return a.Status == AccountStatusActive
}

// ChangePassword 修改密码。
// 前置：Status == ACTIVE
func (a *Account) ChangePassword(newHashedPassword string) error {
	if a.Status != AccountStatusActive {
		return errors.ErrInvalidStatusTransition
	}
	if newHashedPassword == "" {
		return errors.ErrInvalidPassword
	}
	a.PasswordHash = newHashedPassword
	a.UpdatedAt = time.Now()
	return nil
}

// IsAdmin 是否管理员。
func (a *Account) IsAdmin() bool {
	return a.Role == "admin"
}
