package repository

import (
	"context"
)

// OAuthUserInfo 第三方 OAuth 用户信息。
type OAuthUserInfo struct {
	ProviderUserID string
	Email          string
	DisplayName    string
	AvatarURL      string
	RawJSON        string
}

// OAuthProvider 第三方 OAuth 提供方接口。
type OAuthProvider interface {
	GetAuthURL(state, redirectURI string) string
	ExchangeCode(ctx context.Context, code string) (*OAuthUserInfo, error)
}
