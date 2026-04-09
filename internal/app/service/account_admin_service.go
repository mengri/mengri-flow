package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/repository"
	"mengri-flow/internal/infra/config"

	"github.com/google/uuid"
)

// AccountAdminServiceImpl 管理员账号管理服务实现。
type AccountAdminServiceImpl struct {
	accountRepo  repository.AccountRepository         `autowired:""`
	tokenRepo    repository.ActivationTokenRepository `autowired:""`
	identityRepo repository.IdentityRepository        `autowired:""`
	auditRepo    repository.AuditEventRepository      `autowired:""`
	emailSender  repository.IEmailSender              `autowired:""`
	txManager    repository.TransactionManager        `autowired:""`
	cfg          *config.Config                       `autowired:""`
}

var _ IAccountAdminService = (*AccountAdminServiceImpl)(nil)

// CreateAccount 管理员创建账号。
func (s *AccountAdminServiceImpl) CreateAccount(ctx context.Context, req *dto.CreateAccountRequest, operatorID string) (*dto.AccountResponse, error) {
	// 1. 创建 Account 聚合根（业务校验在实体内）
	account, err := entity.NewAccount(req.Email, req.Username, req.DisplayName)
	if err != nil {
		return nil, err
	}
	account.ID = uuid.New().String()

	// 2. 生成激活令牌
	rawToken, err := generateRawToken()
	if err != nil {
		return nil, fmt.Errorf("generate activation token: %w", err)
	}
	tokenTTL := time.Duration(s.cfg.Auth.Activation.TokenExpiry) * time.Second
	activationToken := entity.NewActivationToken(account.ID, rawToken, tokenTTL)

	// 3. 事务中写入 DB
	err = s.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := s.accountRepo.Create(txCtx, account); err != nil {
			return err
		}
		if err := s.tokenRepo.Create(txCtx, activationToken); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 4. 事务外发送激活邮件（可重试）
	activationLink := fmt.Sprintf("%s?token=%s", s.cfg.Email.Activation.BaseURL, rawToken)
	go func() {
		if emailErr := s.emailSender.SendActivationEmail(context.Background(), req.Email, activationLink); emailErr != nil {
			slog.Error("failed to send activation email", "email", req.Email, "error", emailErr)
		}
	}()

	// 5. 审计日志
	s.writeAudit(ctx, operatorID, account.ID, entity.AuditAccountCreated)

	return toAccountResponse(account, &activationToken.ExpiresAt), nil
}

// GetAccountDetail 获取账号详情（含身份列表）。
func (s *AccountAdminServiceImpl) GetAccountDetail(ctx context.Context, accountID string) (*dto.AccountDetailResponse, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	identities, err := s.identityRepo.ListByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	identityBriefs := make([]dto.IdentityBrief, 0, len(identities))
	for _, id := range identities {
		identityBriefs = append(identityBriefs, dto.IdentityBrief{
			IdentityID: id.ID,
			LoginType:  string(id.LoginType),
			BoundAt:    id.CreatedAt.Format(time.RFC3339),
		})
	}

	return &dto.AccountDetailResponse{
		AccountResponse: *toAccountResponse(account, nil),
		Identities:      identityBriefs,
	}, nil
}

// ListAccounts 分页查询账号列表。
func (s *AccountAdminServiceImpl) ListAccounts(ctx context.Context, req *dto.ListAccountsRequest) (*dto.ListAccountsResponse, error) {
	page, pageSize := normalizePageParams(req.Page, req.PageSize)
	offset := (page - 1) * pageSize

	var statusFilter *entity.AccountStatus
	if req.Status != "" {
		st := entity.AccountStatus(req.Status)
		statusFilter = &st
	}

	accounts, total, err := s.accountRepo.List(ctx, offset, pageSize, statusFilter, req.Keyword)
	if err != nil {
		return nil, err
	}

	items := make([]dto.AccountResponse, 0, len(accounts))
	for _, a := range accounts {
		items = append(items, *toAccountResponse(a, nil))
	}

	return &dto.ListAccountsResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ChangeAccountStatus 管理员变更账号状态。
func (s *AccountAdminServiceImpl) ChangeAccountStatus(ctx context.Context, accountID string, req *dto.ChangeStatusRequest, operatorID string) (*dto.AccountResponse, error) {
	var account *entity.Account

	err := s.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		var err error
		account, err = s.accountRepo.GetByID(txCtx, accountID)
		if err != nil {
			return err
		}

		// 执行状态迁移
		switch req.Action {
		case "lock":
			err = account.Lock()
		case "unlock":
			err = account.Unlock()
		case "disable":
			err = account.Disable()
		case "enable":
			err = account.Enable()
		default:
			return domainErr.ErrInvalidStatusTransition
		}
		if err != nil {
			return err
		}

		return s.accountRepo.Update(txCtx, account)
	})
	if err != nil {
		return nil, err
	}

	// 审计日志
	eventType := statusActionToAuditEvent(req.Action)
	s.writeAudit(ctx, operatorID, accountID, eventType)

	return toAccountResponse(account, nil), nil
}

// ResendActivation 重发激活邮件。
func (s *AccountAdminServiceImpl) ResendActivation(ctx context.Context, accountID string, reason string, operatorID string) (*dto.ResendActivationResponse, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	if account.Status != entity.AccountStatusPendingActivation {
		return nil, domainErr.ErrAlreadyActivated
	}

	// 作废旧 token
	if err := s.tokenRepo.InvalidateByAccountID(ctx, accountID); err != nil {
		return nil, fmt.Errorf("invalidate old tokens: %w", err)
	}

	// 生成新 token
	rawToken, err := generateRawToken()
	if err != nil {
		return nil, fmt.Errorf("generate activation token: %w", err)
	}
	tokenTTL := time.Duration(s.cfg.Auth.Activation.TokenExpiry) * time.Second
	activationToken := entity.NewActivationToken(accountID, rawToken, tokenTTL)

	if err := s.tokenRepo.Create(ctx, activationToken); err != nil {
		return nil, err
	}

	// 发送激活邮件
	activationLink := fmt.Sprintf("%s?token=%s", s.cfg.Email.Activation.BaseURL, rawToken)
	go func() {
		if emailErr := s.emailSender.SendActivationEmail(context.Background(), account.Email.String(), activationLink); emailErr != nil {
			slog.Error("failed to resend activation email", "email", account.Email.String(), "error", emailErr)
		}
	}()

	// 审计日志
	s.writeAudit(ctx, operatorID, accountID, entity.AuditActivationEmailResent)

	return &dto.ResendActivationResponse{
		Sent:               true,
		ActivationExpireAt: activationToken.ExpiresAt.Format(time.RFC3339),
		ThrottleSec:        s.cfg.Auth.Activation.ResendCooldown,
	}, nil
}

// ListAuditEvents 查询审计事件列表。
func (s *AccountAdminServiceImpl) ListAuditEvents(ctx context.Context, req *dto.AuditEventFilter) (*dto.AuditEventListResponse, error) {
	page, pageSize := normalizePageParams(req.Page, req.PageSize)
	offset := (page - 1) * pageSize

	filter := repository.AuditFilter{
		AccountID: req.AccountID,
		EventType: req.EventType,
		Offset:    offset,
		Limit:     pageSize,
	}

	if req.From != "" {
		t, err := time.Parse(time.RFC3339, req.From)
		if err == nil {
			filter.From = &t
		}
	}
	if req.To != "" {
		t, err := time.Parse(time.RFC3339, req.To)
		if err == nil {
			filter.To = &t
		}
	}

	events, total, err := s.auditRepo.ListByFilter(ctx, filter)
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

// --- 私有方法 ---

func (s *AccountAdminServiceImpl) writeAudit(ctx context.Context, actorID, targetID, eventType string) {
	audit, _ := entity.NewAuditEvent(actorID, targetID, eventType, entity.AuditResultSuccess, "", "")
	if audit != nil {
		audit.ID = uuid.New().String()
		if err := s.auditRepo.Create(ctx, audit); err != nil {
			slog.Error("failed to write audit event", "type", eventType, "error", err)
		}
	}
}

func toAccountResponse(account *entity.Account, activationExpireAt *time.Time) *dto.AccountResponse {
	resp := &dto.AccountResponse{
		AccountID:   account.ID,
		Email:       account.Email.String(),
		Username:    account.Username,
		DisplayName: account.DisplayName,
		Status:      string(account.Status),
		Role:        account.Role,
		CreatedAt:   account.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   account.UpdatedAt.Format(time.RFC3339),
	}
	if account.ActivatedAt != nil {
		resp.ActivatedAt = account.ActivatedAt.Format(time.RFC3339)
	}
	if activationExpireAt != nil {
		resp.ActivationExpireAt = activationExpireAt.Format(time.RFC3339)
	}
	return resp
}

func normalizePageParams(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func generateRawToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func statusActionToAuditEvent(action string) string {
	switch action {
	case "lock":
		return entity.AuditAccountLocked
	case "unlock":
		return entity.AuditAccountUnlocked
	case "disable":
		return entity.AuditAccountDisabled
	case "enable":
		return entity.AuditAccountEnabled
	default:
		return ""
	}
}
