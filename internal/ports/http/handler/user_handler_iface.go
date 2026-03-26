package handler

import (
	"mengri-flow/pkg/autowire"

	"github.com/gin-gonic/gin"
)

// UserHandler 定义用户 HTTP 处理器接口。
// 接口定义在 Ports 层（调用方），Router 依赖此接口而非具体实现。
type UserHandler interface {
	Create(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func init() {
	autowire.Auto(func() UserHandler {
		return &UserHandlerImpl{}
	})
}
