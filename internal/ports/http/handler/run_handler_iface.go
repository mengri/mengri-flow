package handler

import (
	"mengri-flow/pkg/autowire"

	"github.com/gin-gonic/gin"
)

// RunHandler 运行记录 HTTP 处理器接口
type IRunHandler interface {
	ListRuns(c *gin.Context)
	GetRunDetail(c *gin.Context)
	GetExecutionTimeline(c *gin.Context)
	RetryRun(c *gin.Context)
	GetRunStats(c *gin.Context)
}

func init() {
	autowire.Auto(func() IRunHandler {
		return &RunHandlerImpl{}
	})
}
