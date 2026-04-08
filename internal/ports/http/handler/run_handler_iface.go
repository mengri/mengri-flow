package handler

import "github.com/gin-gonic/gin"

// IRunHandler 运行记录处理器接口
type IRunHandler interface {
	ListRuns(c *gin.Context)
	GetRunDetail(c *gin.Context)
	GetExecutionTimeline(c *gin.Context)
	RetryRun(c *gin.Context)
	GetRunStats(c *gin.Context)
}

// Ensure RunHandler implements IRunHandler
var _ IRunHandler = (*RunHandler)(nil)
