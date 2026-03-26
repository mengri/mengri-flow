package repository

import (
	"context"
)

// TransactionManager 事务管理器接口。
type TransactionManager interface {
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
