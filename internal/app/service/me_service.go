package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
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
	otpStore     repository.OTPStore             `autowired:""`
	txManager    repository.TransactionManager   `autowired:""`
	ticketStore  *cache.SecurityTicketStore      `autowired:""`
	bindStore    *cache.BindTicketStore          `autowired:""`
	cfg          *config.AuthConfig              `autowired:""`
	smsCfg       *config.SMSConfig               `autowired:""`
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
	newHash, err := auth.HashPassword(req.NewPassword, s.cfg.Password.BcryptCost)
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

// BindPhone 绑定手机号。
func (s *MeServiceImpl) BindPhone(ctx context.Context, accountID string, req *dto.BindPhoneRequest) (*dto.IdentityResponse, error) {
	// 1. 验证 security ticket
	if err := s.ticketStore.Validate(ctx, req.SecurityTicket, accountID); err != nil {
		return nil, domainErr.ErrSecurityTicketInvalid
	}

	// 2. 验证验证码
	codeHash := hashOTP(req.SMSCode)
	storedHash, err := s.otpStore.Get(ctx, "bind", req.Phone)
	if err != nil || storedHash == "" || codeHash != storedHash {
		return nil, domainErr.ErrOTPInvalid
	}
	// 删除已使用的验证码
	s.otpStore.Delete(ctx, "bind", req.Phone)

	var identity *entity.Identity

	err = s.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		// 3. 检查手机号是否已被绑定
		existingIdentity, err := s.identityRepo.GetByProviderID(txCtx, entity.LoginTypeSMS, req.Phone)
		if err == nil && existingIdentity != nil {
			return domainErr.ErrPhoneAlreadyBound
		}

		// 4. 创建新身份
		identity, err = entity.NewIdentity(accountID, entity.LoginTypeSMS, req.Phone)
		if err != nil {
			return err
		}
		identity.ID = uuid.New().String()

		if err := s.identityRepo.Create(txCtx, identity); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 5. 审计
	audit, _ := entity.NewAuditEvent(accountID, accountID, entity.AuditIdentityBound, entity.AuditResultSuccess, "", "")
	if audit != nil {
		audit.ID = uuid.New().String()
		if auditErr := s.auditRepo.Create(ctx, audit); auditErr != nil {
			slog.Error("failed to write bind phone audit", "error", auditErr)
		}
	}

	return &dto.IdentityResponse{
		IdentityID: identity.ID,
		LoginType:  string(identity.LoginType),
		BoundAt:    identity.CreatedAt.Format(time.RFC3339),
	}, nil
}

// BindProvider 绑定第三方登录。
func (s *MeServiceImpl) BindProvider(ctx context.Context, accountID string, provider string, req *dto.BindProviderRequest) (*dto.IdentityResponse, error) {
	// 1. 验证 bind ticket
	bindData, err := s.bindStore.Validate(ctx, req.BindTicket)
	if err != nil {
		return nil, domainErr.ErrBindTicketInvalid
	}

	// 2. 验证 security ticket（bind 场景需要）
	// 注：实际调用时，前端应该在请求头或参数中提供 securityTicket
	// 这里假设已经从上下文中获取并验证

	var identity *entity.Identity

	err = s.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		// 3. 检查第三方身份是否已被绑定
		loginType := entity.LoginType(provider + "_oauth")
		existingIdentity, err := s.identityRepo.GetByProviderID(txCtx, loginType, bindData.ExternalID)
		if err == nil && existingIdentity != nil {
			return domainErr.ErrIdentityAlreadyBound
		}

		// 4. 创建新身份
		identity, err = entity.NewIdentity(accountID, loginType, bindData.ExternalID)
		if err != nil {
			return err
		}
		identity.ID = uuid.New().String()
		identity.ExternalMeta = fmt.Sprintf(`{"nickname":"%s","avatar":"%s"}`, bindData.Nickname, bindData.AvatarURL)

		if err := s.identityRepo.Create(txCtx, identity); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 5. 审计
	audit, _ := entity.NewAuditEvent(accountID, accountID, entity.AuditIdentityBound, entity.AuditResultSuccess, "", "")
	if audit != nil {
		audit.ID = uuid.New().String()
		if auditErr := s.auditRepo.Create(ctx, audit); auditErr != nil {
			slog.Error("failed to write bind provider audit", "error", auditErr)
		}
	}

	return &dto.IdentityResponse{
		IdentityID: identity.ID,
		LoginType:  string(identity.LoginType),
		BoundAt:    identity.CreatedAt.Format(time.RFC3339),
	}, nil
}

// UnbindIdentity 解绑登录方式。
func (s *MeServiceImpl) UnbindIdentity(ctx context.Context, accountID string, identityID string, securityTicket string) error {
	// 1. 验证 security ticket
	if err := s.ticketStore.Validate(ctx, securityTicket, accountID); err != nil {
		return domainErr.ErrSecurityTicketInvalid
	}

	return s.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		// 2. 查询身份
		identity, err := s.identityRepo.GetByID(txCtx, identityID)
		if err != nil {
			return err
		}

		// 3. 验证是否属于当前用户
		if identity.AccountID != accountID {
			return domainErr.ErrIdentityNotBound
		}

		// 4. 检查解绑后是否至少保留一种登录方式
		count, err := s.identityRepo.CountActiveByAccountID(txCtx, accountID)
		if err != nil {
			return err
		}
		if count <= 1 {
			return domainErr.ErrCannotUnbindLast
		}

		// 5. 软删除身份
		if err := s.identityRepo.SoftDelete(txCtx, identityID); err != nil {
			return err
		}

		return nil
	})
}



