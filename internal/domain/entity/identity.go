package entity

import (
	"mengri-flow/internal/domain/errors"
	"time"
)

// LoginType 登录方式枚举
type LoginType string

const (
	LoginTypePassword LoginType = "password"
	LoginTypeSMS      LoginType = "sms"
	LoginTypeWechatQR LoginType = "wechat_qr"
	LoginTypeLarkQR   LoginType = "lark_qr"
	LoginTypeGithub   LoginType = "github_oauth"
)

// Identity 登录身份（Account 的子实体）。
type Identity struct {
	ID           string
	AccountID    string
	LoginType    LoginType
	ExternalID   string // 密码登录为 email，第三方为 provider user id
	ExternalMeta string // JSON，存昵称、头像等
	CreatedAt    time.Time
	DeletedAt    *time.Time
}

// NewIdentity 创建身份绑定记录。
func NewIdentity(accountID string, loginType LoginType, externalID string) (*Identity, error) {
	if accountID == "" {
		return nil, errors.ErrNotFound
	}
	if externalID == "" {
		return nil, errors.ErrIdentityNotBound
	}

	return &Identity{
		AccountID:  accountID,
		LoginType:  loginType,
		ExternalID: externalID,
		CreatedAt:  time.Now(),
	}, nil
}

// CanUnbind 检查是否可以解绑（剩余可用登录方式 > 1）。
func CanUnbind(totalActiveCount int) bool {
	return totalActiveCount > 1
}
