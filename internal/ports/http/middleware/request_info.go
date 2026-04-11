package middleware

import (
	"mengri-flow/pkg/ctxutil"

	"github.com/gin-gonic/gin"
)

// RequestInfo 将客户端 IP 和 User-Agent 注入 context.Context，
// 供下游 service 层通过 ctxutil.ClientIP / ctxutil.UserAgent 获取。
func RequestInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx = ctxutil.WithClientIP(ctx, c.ClientIP())
		ctx = ctxutil.WithUserAgent(ctx, c.GetHeader("User-Agent"))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
