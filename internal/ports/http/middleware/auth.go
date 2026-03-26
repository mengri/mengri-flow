package middleware

import (
	"strings"

	"mengri-flow/internal/infra/auth"
	"mengri-flow/pkg/response"

	"github.com/gin-gonic/gin"
)

// Auth JWT 认证中间件。
// 从 Authorization 头提取 Bearer token，解析并注入 accountId、role 到 gin.Context。
func Auth(jwtMgr *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorized(c, "session expired or invalid")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, "session expired or invalid")
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims, err := jwtMgr.ParseToken(tokenStr)
		if err != nil {
			response.Unauthorized(c, "session expired or invalid")
			c.Abort()
			return
		}

		c.Set("accountId", claims.AccountID)
		c.Set("role", claims.Role)
		c.Next()
	}
}
