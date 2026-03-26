package service

import (
	"context"
	"mengri-flow/internal/app/dto"
	"mengri-flow/pkg/autowire"
)

// AccountAdminService 管理员账号管理服务接口。
type AccountAdminService interface {
	CreateAccount(ctx context.Context, req *dto.CreateAccountRequest, operatorID string) (*dto.AccountResponse, error)
	GetAccountDetail(ctx context.Context, accountID string) (*dto.AccountDetailResponse, error)
	ListAccounts(ctx context.Context, req *dto.ListAccountsRequest) (*dto.ListAccountsResponse, error)
	ChangeAccountStatus(ctx context.Context, accountID string, req *dto.ChangeStatusRequest, operatorID string) (*dto.AccountResponse, error)
	ResendActivation(ctx context.Context, accountID string, reason string, operatorID string) (*dto.ResendActivationResponse, error)
	ListAuditEvents(ctx context.Context, req *dto.AuditEventFilter) (*dto.AuditEventListResponse, error)
}

func init() {
	autowire.Auto(func() AccountAdminService {
		return &AccountAdminServiceImpl{}
	})
}
