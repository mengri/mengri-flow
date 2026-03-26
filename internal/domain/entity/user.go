package entity

import (
	"mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/valueobject"
	"time"
)

// User 是用户聚合根，业务逻辑必须留在 Entity 方法中（拒绝贫血模型）。
type User struct {
	ID        uint64
	Username  string
	Email     valueobject.Email
	Password  string // 已哈希的密码
	Status    UserStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserStatus int

const (
	UserStatusActive   UserStatus = 1
	UserStatusInactive UserStatus = 2
	UserStatusBanned   UserStatus = 3
)

// NewUser 创建一个新用户实体，包含业务校验逻辑。
func NewUser(username string, email string, hashedPassword string) (*User, error) {
	if username == "" {
		return nil, errors.ErrInvalidUsername
	}
	if len(username) < 2 || len(username) > 50 {
		return nil, errors.ErrInvalidUsername
	}

	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		Username:  username,
		Email:     emailVO,
		Password:  hashedPassword,
		Status:    UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Activate 激活用户
func (u *User) Activate() error {
	if u.Status == UserStatusBanned {
		return errors.ErrUserBanned
	}
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now()
	return nil
}

// Ban 封禁用户
func (u *User) Ban() {
	u.Status = UserStatusBanned
	u.UpdatedAt = time.Now()
}

// IsActive 判断用户是否活跃
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// ChangeEmail 修改邮箱（包含业务校验）
func (u *User) ChangeEmail(newEmail string) error {
	emailVO, err := valueobject.NewEmail(newEmail)
	if err != nil {
		return err
	}
	u.Email = emailVO
	u.UpdatedAt = time.Now()
	return nil
}
