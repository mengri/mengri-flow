package repository

import (
	"context"
)

// CredentialRepository 账号凭据（密码）仓储接口。
type CredentialRepository interface {
	Create(ctx context.Context, accountID string, passwordHash string) error
	GetByAccountID(ctx context.Context, accountID string) (passwordHash string, err error)
	UpdatePassword(ctx context.Context, accountID string, passwordHash string) error
}
