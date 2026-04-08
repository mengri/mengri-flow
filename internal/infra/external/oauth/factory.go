package oauth

import (
	"mengri-flow/internal/domain/repository"
	"mengri-flow/internal/infra/config"
	"mengri-flow/pkg/autowire"
)

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
	if len(providers) > 0 {
		autowire.Auto(func() map[string]repository.OAuthProvider {
			return providers
		})
	}
}
