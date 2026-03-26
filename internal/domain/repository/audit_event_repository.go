package repository

import (
	"context"
	"mengri-flow/internal/domain/entity"
	"time"
)

// AuditFilter 审计事件查询过滤器。
type AuditFilter struct {
	AccountID string
	EventType string
	From      *time.Time
	To        *time.Time
	Offset    int
	Limit     int
}

// AuditEventRepository 审计事件仓储接口。
type AuditEventRepository interface {
	Create(ctx context.Context, event *entity.AuditEvent) error
	ListByAccountID(ctx context.Context, accountID string, offset, limit int) ([]*entity.AuditEvent, int64, error)
	ListByFilter(ctx context.Context, filter AuditFilter) ([]*entity.AuditEvent, int64, error)
}
