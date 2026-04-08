package service

import (
	"context"
	"mengri-flow/internal/app/dto"
	"mengri-flow/pkg/autowire"
)

// AuthService 认证应用服务接口。
type AuthService interface {
	// 激活流程
	ValidateActivationToken(ctx context.Context, token string) (*dto.ActivationValidateResponse, error)
	ConfirmActivation(ctx context.Context, req *dto.ActivationConfirmRequest) (*dto.ActivationConfirmResponse, error)

	// 登录
	LoginByPassword(ctx context.Context, req *dto.PasswordLoginRequest) (*dto.LoginResponse, error)
	SendSMSCode(ctx context.Context, req *dto.SMSSendRequest) (*dto.SMSSendResponse, error)
	LoginBySMS(ctx context.Context, req *dto.SMSLoginRequest) (*dto.LoginResponse, error)

	// OAuth
	GetOAuthURL(ctx context.Context, provider, scene, redirectURI string) (*dto.OAuthURLResponse, error)
	HandleOAuthCallback(ctx context.Context, provider, code, state string) (*dto.OAuthCallbackResponse, error)

	// Token
	RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error)
	Logout(ctx context.Context, accountID, refreshTokenHash string) error
}

func init() {
	autowire.Auto(func() AuthService {
		return &AuthServiceImpl{}
	})
}
