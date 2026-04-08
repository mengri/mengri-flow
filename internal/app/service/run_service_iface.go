package service

import (
	"context"

	"mengri-flow/internal/app/dto"
)

type IRunService interface {
	ListRuns(ctx context.Context, req *dto.ListRunsRequest) (*dto.ListRunsResponse, error)
	GetRunDetail(ctx context.Context, id string) (*dto.RunDetailResponse, error)
	GetExecutionTimeline(ctx context.Context, runID string) ([]dto.ExecutionTimelineResponse, error)
	RetryRun(ctx context.Context, runID string, operatorID string) (*dto.RunResponse, error)
	GetRunStats(ctx context.Context) (*dto.RunStatsResponse, error)
}

var _ IRunService = (*RunService)(nil)
