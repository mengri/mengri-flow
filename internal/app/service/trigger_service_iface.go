package service

import (
	"context"

	"mengri-flow/internal/app/dto"
	"mengri-flow/pkg/autowire"
)

type ITriggerService interface {
	CreateTrigger(ctx context.Context, req *dto.CreateTriggerRequest, creatorID string) (*dto.TriggerResponse, error)
	ListTriggers(ctx context.Context, req *dto.ListTriggersRequest) (*dto.ListTriggersResponse, error)
	GetTrigger(ctx context.Context, id string) (*dto.TriggerResponse, error)
	UpdateTrigger(ctx context.Context, id string, req *dto.UpdateTriggerRequest, operatorID string) (*dto.TriggerResponse, error)
	DeleteTrigger(ctx context.Context, id string, operatorID string) error
	EnableTrigger(ctx context.Context, id string, operatorID string) error
	DisableTrigger(ctx context.Context, id string, operatorID string) error
	PublishToCluster(ctx context.Context, triggerID string, clusterID string, operatorID string) error
}

var _ ITriggerService = (*triggerServiceImpl)(nil)

func init() {
	autowire.Auto(func() ITriggerService {
		return new(triggerServiceImpl)
	})
}
