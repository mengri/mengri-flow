package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"mengri-flow/internal/domain/repository"
	"mengri-flow/internal/infra/config"
)

// GitHubProvider GitHub OAuth 提供商
type GitHubProvider struct {
	config *oauth2.Config
}

// NewGitHubProvider 创建 GitHub OAuth 提供商
func NewGitHubProvider(cfg *config.OAuthProviderConf) *GitHubProvider {
	return &GitHubProvider{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURI,
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		},
	}
}

// GetAuthURL 获取授权地址
func (p *GitHubProvider) GetAuthURL(state, redirectURI string) string {
	if redirectURI != "" {
		p.config.RedirectURL = redirectURI
	}
	return p.config.AuthCodeURL(state)
}

// ExchangeCode 交换 code 获取用户信息
func (p *GitHubProvider) ExchangeCode(ctx context.Context, code string) (*repository.OAuthUserInfo, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}

	// 获取用户信息
	client := p.config.Client(ctx, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("get user info: %w", err)
	}
	defer resp.Body.Close()

	var user struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode user info: %w", err)
	}

	displayName := user.Name
	if displayName == "" {
		displayName = user.Login
	}

	return &repository.OAuthUserInfo{
		ProviderUserID: fmt.Sprintf("%d", user.ID),
		Email:          user.Email,
		DisplayName:    displayName,
		AvatarURL:      user.AvatarURL,
	}, nil
}
