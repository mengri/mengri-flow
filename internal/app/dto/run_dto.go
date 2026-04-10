package dto

import (
	"time"
)

type RunResponse struct {
	ID          string    `json:"id"`
	FlowID      string    `json:"flowId"`
	FlowName    string    `json:"flowName"`
	TriggerID   string    `json:"triggerId"`
	TriggerType string    `json:"triggerType"`
	Status      string    `json:"status"`
	StartedAt   time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

type RunDetailResponse struct {
	RunResponse
	Input      map[string]interface{}  `json:"input"`
	Output     map[string]interface{}  `json:"output,omitempty"`
	Error      string                  `json:"error,omitempty"`
	NodeCount  int                     `json:"nodeCount"`
	NodeStatus map[string]string       `json:"nodeStatus"`
}

type ExecutionTimelineResponse struct {
	NodeID       string    `json:"nodeId"`
	NodeName     string    `json:"nodeName"`
	ToolID       string    `json:"toolId"`
	ToolName     string    `json:"toolName"`
	Status       string    `json:"status"`
	StartedAt    time.Time `json:"startedAt"`
	CompletedAt  *time.Time `json:"completedAt,omitempty"`
	ExecutionTime int64     `json:"executionTime,omitempty"`
	Error        string    `json:"error,omitempty"`
}

type RetryRunRequest struct {
	RunID string `json:"runId" binding:"required"`
}

type RunStatsResponse struct {
	TotalRuns   int64                `json:"totalRuns"`
	SuccessRuns int64                `json:"successRuns"`
	FailedRuns  int64                `json:"failedRuns"`
	RunningRuns int64                `json:"runningRuns"`
	TodayRuns   int64                `json:"todayRuns"`
	WeekRuns    int64                `json:"weekRuns"`
	MonthRuns   int64                `json:"monthRuns"`
	SuccessRate float64              `json:"successRate"`
	AvgDuration int64                `json:"avgDuration"`
	Trend       []RunStatsTrendItem  `json:"trend"`
}

type RunStatsTrendItem struct {
	Date    string `json:"date"`
	Success int64  `json:"success"`
	Failed  int64  `json:"failed"`
}

type ListRunsRequest struct {
	FlowID    string `json:"flowId" binding:"omitempty"`
	TriggerID string `json:"triggerId" binding:"omitempty"`
	Status    string `json:"status" binding:"omitempty"`
	Page      int    `json:"page" binding:"min=1"`
	PageSize  int    `json:"pageSize" binding:"min=1,max=100"`
}

type ListRunsResponse struct {
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
	List     []RunResponse `json:"list"`
}
