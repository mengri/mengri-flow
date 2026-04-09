package handler

import (
	"mengri-flow/pkg/autowire"

	"github.com/gin-gonic/gin"
)

// AccountAdminHandler 管理员账号管理 HTTP 处理器接口。
type IAccountAdminHandler interface {
	Create(c *gin.Context)
	List(c *gin.Context)
	GetDetail(c *gin.Context)
	ChangeStatus(c *gin.Context)
	ResendActivation(c *gin.Context)
	ListAuditEvents(c *gin.Context)
}

func init() {
	autowire.Auto(func() IAccountAdminHandler {
		return &AccountAdminHandlerImpl{}
	})
}
