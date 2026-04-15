package middleware

import (
	"strings"

	"mengri-flow/internal/infra/auth"
	"mengri-flow/pkg/response"

	"github.com/gin-gonic/gin"
)

// Auth JWT 认证中间件。
// 从 Authorization 头提取 Bearer token，解析并注入 accountID、role 到 gin.Context。
// Refresh Token 不允许用于 API 访问，防止 token 混用。
func Auth(jwtMgr auth.IJWTManager) gin.HandlerFunc {
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

		// Refresh Token 不能用于 API 访问
		if claims.TokenType == auth.TokenTypeRefresh {
			response.Unauthorized(c, "refresh token cannot be used for API access")
			c.Abort()
			return
		}

		c.Set("accountID", claims.AccountID)
		c.Set("role", claims.Role)
		c.Next()
	}
}
