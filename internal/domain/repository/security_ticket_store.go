package repository

import (
	"context"
	"time"
)

type SecurityTicketStore interface {
	Generate(ctx context.Context, accountID string) (string, error)
	Validate(ctx context.Context, ticket, expectedAccountID string) error
	TTL() time.Duration
}
