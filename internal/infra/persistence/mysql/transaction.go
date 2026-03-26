package mysql

import (
	"context"
	"mengri-flow/internal/domain/repository"

	"gorm.io/gorm"
)

// txContextKey 用于在 context 中传递事务 DB。
type txContextKey struct{}

// TransactionManagerImpl 是 TransactionManager 的 GORM 实现。
type TransactionManagerImpl struct {
	db *gorm.DB
}

var _ repository.TransactionManager = (*TransactionManagerImpl)(nil)

// NewTransactionManager 创建事务管理器。
func NewTransactionManager(db *gorm.DB) *TransactionManagerImpl {
	return &TransactionManagerImpl{db: db}
}

// RunInTransaction 在事务中执行 fn。
// 如果 fn 返回 error，事务自动回滚；否则自动提交。
// 支持嵌套调用：如果 ctx 中已有事务 DB，则复用它（不开启新事务）。
func (m *TransactionManagerImpl) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// 如果上层已经开启了事务，直接复用
	if _, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok {
		return fn(ctx)
	}

	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txContextKey{}, tx)
		return fn(txCtx)
	})
}

// DBFromContext 从 context 中获取事务 DB；如果不在事务中，返回原始 DB。
// 供各 Repository 实现使用，以支持事务内操作。
func DBFromContext(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return fallback.WithContext(ctx)
}
