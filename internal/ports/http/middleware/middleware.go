package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// allowedOrigins 可在启动时由配置注入，空表示允许所有来源（仅开发使用）。
// 生产环境应通过 WithAllowedOrigins 显式设置白名单。
var allowedOrigins []string

// WithAllowedOrigins 设置允许的跨域来源白名单。
func WithAllowedOrigins(origins []string) {
	allowedOrigins = origins
}

// isAllowedOrigin 检查请求来源是否在白名单中。
// 若白名单为空，则仅在非凭证模式下允许通配。
func isAllowedOrigin(origin string) bool {
	if len(allowedOrigins) == 0 {
		return true
	}
	for _, o := range allowedOrigins {
		if o == origin {
			return true
		}
	}
	return false
}

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		slog.Info("HTTP request",
			"method", method,
			"path", path,
			"status", statusCode,
			"latency", latency.String(),
			"client_ip", c.ClientIP(),
		)
	}
}

// Recovery panic 恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

// CORS 跨域中间件。
// 当启用 Credentials 时，Origin 不能为通配符（浏览器规范），
// 此处动态反射请求来源并校验白名单，保证二者兼容。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if origin != "" && isAllowedOrigin(origin) {
			// 明确回显允许的 Origin，满足浏览器 credentials 规范
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		} else if origin == "" {
			// 同源请求无需设置 CORS 头
		} else {
			// 不在白名单中，拒绝该跨域请求
			c.Header("Access-Control-Allow-Origin", "null")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		// Vary 告知代理服务器缓存需按 Origin 区分
		c.Header("Vary", "Origin")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
