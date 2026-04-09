package oauth

import (
	"mengri-flow/internal/domain/repository"
	"mengri-flow/internal/infra/config"
	"mengri-flow/pkg/autowire"
)

type IOAuthProviders interface {
	GetProvider(provider string) (repository.OAuthProvider, bool)
}
type OAuthProviders struct {
	providers map[string]repository.OAuthProvider
}

// GetProvider implements [IOAuthProviders].
func (o *OAuthProviders) GetProvider(provider string) (repository.OAuthProvider, bool) {
	p, ok := o.providers[provider]
	return p, ok
}

// InitOAuthProviders 初始化并注册所有 OAuth 提供商
func InitOAuthProviders(oauthCfg *config.OAuthConfig) {
	providers := make(map[string]repository.OAuthProvider)

	// GitHub
	if oauthCfg.GitHub.ClientID != "" && oauthCfg.GitHub.ClientSecret != "" {
		providers["github"] = NewGitHubProvider(&oauthCfg.GitHub)
	}

	// 可以在这里添加更多提供商（微信、飞书等）
	// if oauthCfg.WeChat.AppID != "" && oauthCfg.WeChat.AppSecret != "" {
	//     providers["wechat"] = NewWeChatProvider(&oauthCfg.WeChat)
	// }
	// if oauthCfg.Lark.AppID != "" && oauthCfg.Lark.AppSecret != "" {
	//     providers["lark"] = NewLarkProvider(&oauthCfg.Lark)
	// }

	// 注册到 autowire

	autowire.Auto(func() IOAuthProviders {
		return &OAuthProviders{
			providers: providers,
		}
	})

}
