package service

import (
	"context"
	"mengri-flow/internal/app/dto"
	"mengri-flow/pkg/autowire"
)

// MeService 账号中心服务接口（用户自助操作）。
type IMeService interface {
	GetProfile(ctx context.Context, accountID string) (*dto.ProfileResponse, error)
	ListIdentities(ctx context.Context, accountID string) (*dto.IdentityListResponse, error)
	ChangePassword(ctx context.Context, accountID string, req *dto.ChangePasswordRequest) (*dto.ChangePasswordResponse, error)
	SecurityVerify(ctx context.Context, accountID string, password string) (*dto.SecurityTicketResponse, error)
	LoginHistory(ctx context.Context, accountID string, page, pageSize int) (*dto.AuditEventListResponse, error)
}

func init() {
	autowire.Auto(func() IMeService {
		return &meServiceImpl{}
	})
}
