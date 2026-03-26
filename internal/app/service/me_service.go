package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/domain/entity"
	"mengri-flow/internal/domain/repository"
	"mengri-flow/internal/infra/auth"
	"mengri-flow/internal/infra/cache"
	"mengri-flow/internal/infra/config"

	"github.com/google/uuid"
)

// MeServiceImpl 账号中心服务实现。
type MeServiceImpl struct {
	accountRepo  repository.AccountRepository    `autowired:""`
	credRepo     repository.CredentialRepository `autowired:""`
	identityRepo repository.IdentityRepository   `autowired:""`
	sessionStore repository.SessionStore         `autowired:""`
	auditRepo    repository.AuditEventRepository `autowired:""`
	txManager    repository.TransactionManager   `autowired:""`
	ticketStore  *cache.SecurityTicketStore      `autowired:""`
	cfg          *config.Config                  `autowired:""`
}

var _ MeService = (*MeServiceImpl)(nil)

// GetProfile 获取当前用户资料。
func (s *MeServiceImpl) GetProfile(ctx context.Context, accountID string) (*dto.ProfileResponse, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	return &dto.ProfileResponse{
		AccountID:     account.ID,
		Email:         account.Email.String(),
		Username:      account.Username,
		DisplayName:   account.DisplayName,
		AccountStatus: string(account.Status),
		Role:          account.Role,
	}, nil
}

// ListIdentities 列出当前用户所有登录身份。
func (s *MeServiceImpl) ListIdentities(ctx context.Context, accountID string) (*dto.IdentityListResponse, error) {
	identities, err := s.identityRepo.ListByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	briefs := make([]dto.IdentityBrief, 0, len(identities))
	for _, id := range identities {
		briefs = append(briefs, dto.IdentityBrief{
			IdentityID: id.ID,
			LoginType:  string(id.LoginType),
			BoundAt:    id.CreatedAt.Format(time.RFC3339),
		})
	}

	canUnbind := len(identities) > 1

	return &dto.IdentityListResponse{
		Identities:    briefs,
		CanUnbindLast: canUnbind,
	}, nil
}

// ChangePassword 修改密码。
func (s *MeServiceImpl) ChangePassword(ctx context.Context, accountID string, req *dto.ChangePasswordRequest) (*dto.ChangePasswordResponse, error) {
	// 1. 验证旧密码
	storedHash, err := s.credRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if !auth.VerifyPassword(req.OldPassword, storedHash) {
		return nil, fmt.Errorf("old password incorrect")
	}

	// 2. 哈希新密码
	newHash, err := auth.HashPassword(req.NewPassword, s.cfg.Auth.Password.BcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// 3. 更新密码
	revokedCount := 0
	err = s.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := s.credRepo.UpdatePassword(txCtx, accountID, newHash); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 4. 可选：吊销其他 session（事务外，操作 Redis/DB）
	if req.RevokeOtherSessions {
		count, revokeErr := s.sessionStore.RevokeAllByAccountID(ctx, accountID, "")
		if revokeErr != nil {
			slog.Error("failed to revoke other sessions", "error", revokeErr)
		} else {
			revokedCount = count
		}
	}

	// 5. 审计
	audit, _ := entity.NewAuditEvent(accountID, accountID, entity.AuditPasswordChanged, entity.AuditResultSuccess, "", "")
	if audit != nil {
		audit.ID = uuid.New().String()
		if auditErr := s.auditRepo.Create(ctx, audit); auditErr != nil {
			slog.Error("failed to write password change audit", "error", auditErr)
		}
	}

	return &dto.ChangePasswordResponse{
		Changed:         true,
		RevokedSessions: revokedCount,
	}, nil
}

// SecurityVerify 二次安全验证：验证密码后签发票据。
func (s *MeServiceImpl) SecurityVerify(ctx context.Context, accountID string, password string) (*dto.SecurityTicketResponse, error) {
	// 1. 验证密码
	storedHash, err := s.credRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if !auth.VerifyPassword(password, storedHash) {
		return nil, fmt.Errorf("password incorrect")
	}

	// 2. 生成票据
	ticket, err := s.ticketStore.Generate(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// 3. 审计
	audit, _ := entity.NewAuditEvent(accountID, accountID, entity.AuditSecurityVerified, entity.AuditResultSuccess, "", "")
	if audit != nil {
		audit.ID = uuid.New().String()
		if auditErr := s.auditRepo.Create(ctx, audit); auditErr != nil {
			slog.Error("failed to write security verify audit", "error", auditErr)
		}
	}

	expireAt := time.Now().Add(s.ticketStore.TTL())
	return &dto.SecurityTicketResponse{
		SecurityTicket: ticket,
		ExpireAt:       expireAt.Format(time.RFC3339),
		TTLSec:         int(s.ticketStore.TTL().Seconds()),
	}, nil
}

// LoginHistory 查询当前用户登录记录。
func (s *MeServiceImpl) LoginHistory(ctx context.Context, accountID string, page, pageSize int) (*dto.AuditEventListResponse, error) {
	page, pageSize = normalizePageParams(page, pageSize)
	offset := (page - 1) * pageSize

	events, total, err := s.auditRepo.ListByAccountID(ctx, accountID, offset, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]dto.AuditEventItem, 0, len(events))
	for _, e := range events {
		items = append(items, dto.AuditEventItem{
			ID:        e.ID,
			EventType: e.EventType,
			Result:    e.Result,
			IP:        e.IP,
			UA:        e.UA,
			CreatedAt: e.CreatedAt.Format(time.RFC3339),
		})
	}

	return &dto.AuditEventListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
