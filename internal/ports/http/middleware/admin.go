package middleware

import (
	"mengri-flow/pkg/response"

	"github.com/gin-gonic/gin"
)

// Admin 管理员角色检查中间件。
// 依赖 Auth 中间件已注入 "role" 到 gin.Context。
func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			response.Forbidden(c, "no permission")
			c.Abort()
			return
		}
		c.Next()
	}
}
