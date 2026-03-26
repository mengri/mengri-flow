package handler

import (
	"mengri-flow/pkg/autowire"

	"github.com/gin-gonic/gin"
)

// MeHandler 账号中心 HTTP 处理器接口。
type MeHandler interface {
	GetProfile(c *gin.Context)
	ListIdentities(c *gin.Context)
	ChangePassword(c *gin.Context)
	SecurityVerify(c *gin.Context)
	LoginHistory(c *gin.Context)
}

func init() {
	autowire.Auto(func() MeHandler {
		return &MeHandlerImpl{}
	})
}
