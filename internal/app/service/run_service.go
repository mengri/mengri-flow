package service

import (
	"context"
	"fmt"

	"mengri-flow/internal/app/dto"
)

type RunService struct{}

func NewRunService() *RunService {
	return &RunService{}
}

func (s *RunService) ListRuns(ctx context.Context, req *dto.ListRunsRequest) (*dto.ListRunsResponse, error) {
	// TODO: Implement run listing logic
	return &dto.ListRunsResponse{
		Total:    0,
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     []dto.RunResponse{},
	}, nil
}

func (s *RunService) GetRunDetail(ctx context.Context, id string) (*dto.RunDetailResponse, error) {
	// TODO: Implement run detail retrieval logic
	return nil, fmt.Errorf("not implemented")
}

func (s *RunService) GetExecutionTimeline(ctx context.Context, runID string) ([]dto.ExecutionTimelineResponse, error) {
	// TODO: Implement execution timeline retrieval logic
	return []dto.ExecutionTimelineResponse{}, nil
}

func (s *RunService) RetryRun(ctx context.Context, runID string, operatorID string) (*dto.RunResponse, error) {
	// TODO: Implement run retry logic
	return nil, fmt.Errorf("not implemented")
}

func (s *RunService) GetRunStats(ctx context.Context) (*dto.RunStatsResponse, error) {
	// TODO: Implement run stats retrieval logic
	return &dto.RunStatsResponse{
		TotalRuns:   0,
		SuccessRuns: 0,
		FailedRuns:  0,
		RunningRuns: 0,
		TodayRuns:   0,
		WeekRuns:    0,
		MonthRuns:   0,
	}, nil
}
