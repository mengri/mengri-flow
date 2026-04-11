package ctxutil

import "context"

// contextKey 是 unexported 类型，防止外部包冲突。
type contextKey string

const (
	clientIPKey  contextKey = "client_ip"
	userAgentKey contextKey = "user_agent"
)

// WithClientIP 将客户端 IP 注入 context。
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey, ip)
}

// ClientIP 从 context 提取客户端 IP，未设置时返回空字符串。
func ClientIP(ctx context.Context) string {
	v, _ := ctx.Value(clientIPKey).(string)
	return v
}

// WithUserAgent 将 User-Agent 注入 context。
func WithUserAgent(ctx context.Context, ua string) context.Context {
	return context.WithValue(ctx, userAgentKey, ua)
}

// UserAgent 从 context 提取 User-Agent，未设置时返回空字符串。
func UserAgent(ctx context.Context) string {
	v, _ := ctx.Value(userAgentKey).(string)
	return v
}
