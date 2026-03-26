package handler

import (
	"mengri-flow/pkg/autowire"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证 HTTP 处理器接口。
type AuthHandler interface {
	ValidateActivation(c *gin.Context)
	ConfirmActivation(c *gin.Context)
	LoginByPassword(c *gin.Context)
	RefreshToken(c *gin.Context)
	Logout(c *gin.Context)
}

func init() {
	autowire.Auto(func() AuthHandler {
		return &AuthHandlerImpl{}
	})
}
