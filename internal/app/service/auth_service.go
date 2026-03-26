package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/repository"
	"mengri-flow/internal/infra/auth"

	"github.com/google/uuid"
)

// AuthServiceImpl 认证应用服务实现。
type AuthServiceImpl struct {
	accountRepo  repository.AccountRepository         `autowired:""`
	credRepo     repository.CredentialRepository      `autowired:""`
	tokenRepo    repository.ActivationTokenRepository `autowired:""`
	sessionStore repository.SessionStore              `autowired:""`
	auditRepo    repository.AuditEventRepository      `autowired:""`
	identityRepo repository.IdentityRepository        `autowired:""`
	txManager    repository.TransactionManager        `autowired:""`
	jwtManager   *auth.JWTManager                     `autowired:""`
}

var _ AuthService = (*AuthServiceImpl)(nil)

// ValidateActivationToken 验证激活令牌是否有效。
func (s *AuthServiceImpl) ValidateActivationToken(ctx context.Context, rawToken string) (*dto.ActivationValidateResponse, error) {
	tokenHash := hashToken(rawToken)

	token, err := s.tokenRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		return &dto.ActivationValidateResponse{Valid: false}, nil
	}

	account, err := s.accountRepo.GetByID(ctx, token.AccountID)
	if err != nil {
		return &dto.ActivationValidateResponse{Valid: false}, nil
	}

	resp := &dto.ActivationValidateResponse{
		EmailMasked:      maskEmail(account.Email.String()),
		ExpireAt:         token.ExpiresAt.Format(time.RFC3339),
		AlreadyActivated: account.Status != entity.AccountStatusPendingActivation,
	}

	if token.IsUsed() || token.IsExpired() || resp.AlreadyActivated {
		resp.Valid = false
	} else {
		resp.Valid = true
	}

	return resp, nil
}

// ConfirmActivation 确认激活：设置密码并激活账号。
func (s *AuthServiceImpl) ConfirmActivation(ctx context.Context, req *dto.ActivationConfirmRequest) (*dto.ActivationConfirmResponse, error) {
	tokenHash := hashToken(req.Token)

	var resp *dto.ActivationConfirmResponse

	err := s.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		// 1. 查询并校验 token
		token, err := s.tokenRepo.GetByHash(txCtx, tokenHash)
		if err != nil {
			return domainErr.ErrActivationTokenInvalid
		}
		if token.IsUsed() {
			return domainErr.ErrActivationTokenUsed
		}
		if token.IsExpired() {
			return domainErr.ErrActivationTokenExpired
		}

		// 2. 查询账号
		account, err := s.accountRepo.GetByID(txCtx, token.AccountID)
		if err != nil {
			return err
		}

		// 3. 哈希密码
		hashedPwd, err := auth.HashPassword(req.Password, 12)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}

		// 4. 激活账号
		if err := account.Activate(hashedPwd); err != nil {
			return err
		}

		// 5. 更新账号
		if err := s.accountRepo.Update(txCtx, account); err != nil {
			return err
		}

		// 6. 插入凭据
		if err := s.credRepo.Create(txCtx, account.ID, hashedPwd); err != nil {
			return err
		}

		// 7. 创建密码登录身份
		identity, err := entity.NewIdentity(account.ID, entity.LoginTypePassword, account.Email.String())
		if err != nil {
			return err
		}
		identity.ID = uuid.New().String()
		if err := s.identityRepo.Create(txCtx, identity); err != nil {
			return err
		}

		// 8. 标记 token 已使用
		if err := s.tokenRepo.MarkUsed(txCtx, tokenHash); err != nil {
			return err
		}

		resp = &dto.ActivationConfirmResponse{
			Activated:   true,
			AccountID:   account.ID,
			Status:      string(account.Status),
			ActivatedAt: account.ActivatedAt.Format(time.RFC3339),
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 写入审计日志（事务外）
	audit, _ := entity.NewAuditEvent("", resp.AccountID, entity.AuditAccountActivated, entity.AuditResultSuccess, "", "")
	if audit != nil {
		audit.ID = uuid.New().String()
		if auditErr := s.auditRepo.Create(ctx, audit); auditErr != nil {
			slog.Error("failed to write audit event", "error", auditErr)
		}
	}

	return resp, nil
}

// LoginByPassword 密码登录。
func (s *AuthServiceImpl) LoginByPassword(ctx context.Context, req *dto.PasswordLoginRequest) (*dto.LoginResponse, error) {
	// 1. 查找账号：包含 @ 按 email 查询，否则按 username
	var account *entity.Account
	var err error
	if strings.Contains(req.Account, "@") {
		account, err = s.accountRepo.GetByEmail(ctx, req.Account)
	} else {
		account, err = s.accountRepo.GetByUsername(ctx, req.Account)
	}
	if err != nil {
		return nil, domainErr.ErrCredentialsInvalid
	}

	// 2. 检查账号状态
	if account.Status == entity.AccountStatusPendingActivation {
		return nil, domainErr.ErrAccountNotActivated
	}
	if account.Status == entity.AccountStatusLocked {
		return nil, domainErr.ErrAccountLocked
	}
	if account.Status == entity.AccountStatusDisabled {
		return nil, domainErr.ErrAccountDisabled
	}

	// 3. 验证密码
	storedHash, err := s.credRepo.GetByAccountID(ctx, account.ID)
	if err != nil {
		return nil, domainErr.ErrCredentialsInvalid
	}
	if !auth.VerifyPassword(req.Password, storedHash) {
		// 记录登录失败审计
		s.recordLoginAudit(ctx, account.ID, entity.AuditLoginFailed, req.DeviceInfo)
		return nil, domainErr.ErrCredentialsInvalid
	}

	// 4. 签发 Token
	loginResp, err := s.issueTokens(ctx, account, req.DeviceInfo)
	if err != nil {
		return nil, err
	}

	// 5. 记录登录成功审计
	s.recordLoginAudit(ctx, account.ID, entity.AuditLoginSuccess, req.DeviceInfo)

	return loginResp, nil
}

// RefreshToken 刷新 Token。
func (s *AuthServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error) {
	// 1. 解析 refresh token
	claims, err := s.jwtManager.ParseToken(refreshToken)
	if err != nil {
		return nil, domainErr.ErrSessionExpired
	}

	// 2. 验证 session 是否有效
	tokenHash := hashToken(refreshToken)
	accountID, err := s.sessionStore.ValidateRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, domainErr.ErrSessionExpired
	}

	// 3. 吊销旧 refresh token
	if err := s.sessionStore.RevokeRefreshToken(ctx, tokenHash); err != nil {
		slog.Error("failed to revoke old refresh token", "error", err)
	}

	// 4. 查询账号信息
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if !account.CanLogin() {
		return nil, domainErr.ErrAccountLocked
	}

	// 5. 签发新 Token
	loginResp, err := s.issueTokens(ctx, account, dto.DeviceInfo{})
	if err != nil {
		return nil, err
	}

	// 6. 审计
	audit, _ := entity.NewAuditEvent(accountID, accountID, entity.AuditTokenRefreshed, entity.AuditResultSuccess, "", "")
	if audit != nil {
		audit.ID = uuid.New().String()
		if auditErr := s.auditRepo.Create(ctx, audit); auditErr != nil {
			slog.Error("failed to write audit event", "error", auditErr)
		}
	}

	_ = claims // used for parsing validation
	return loginResp, nil
}

// Logout 登出：吊销 refresh token。
func (s *AuthServiceImpl) Logout(ctx context.Context, accountID, refreshTokenRaw string) error {
	if refreshTokenRaw != "" {
		tokenHash := hashToken(refreshTokenRaw)
		if err := s.sessionStore.RevokeRefreshToken(ctx, tokenHash); err != nil {
			slog.Error("failed to revoke refresh token on logout", "error", err)
		}
	}

	// 审计
	audit, _ := entity.NewAuditEvent(accountID, accountID, entity.AuditLogout, entity.AuditResultSuccess, "", "")
	if audit != nil {
		audit.ID = uuid.New().String()
		if auditErr := s.auditRepo.Create(ctx, audit); auditErr != nil {
			slog.Error("failed to write audit event", "error", auditErr)
		}
	}

	return nil
}

// --- 私有方法 ---

// issueTokens 签发 access + refresh token 并存储 session。
func (s *AuthServiceImpl) issueTokens(ctx context.Context, account *entity.Account, device dto.DeviceInfo) (*dto.LoginResponse, error) {
	accessToken, err := s.jwtManager.GenerateAccessToken(account.ID, account.Role)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(account.ID, account.Role)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	// 存储 refresh token session
	sessionID := uuid.New().String()
	refreshHash := hashToken(refreshToken)
	deviceJSON := fmt.Sprintf(`{"ua":"%s","ip":"%s","deviceId":"%s"}`, device.UA, device.IP, device.DeviceID)

	if err := s.sessionStore.SaveRefreshToken(ctx, sessionID, account.ID, refreshHash, deviceJSON, device.IP, s.jwtManager.RefreshTokenExpiry()); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwtManager.AccessTokenExpiry(),
		TokenType:    "Bearer",
		Account: dto.AccountBrief{
			AccountID:   account.ID,
			Email:       account.Email.String(),
			DisplayName: account.DisplayName,
			Status:      string(account.Status),
		},
	}, nil
}

// recordLoginAudit 记录登录审计事件。
func (s *AuthServiceImpl) recordLoginAudit(ctx context.Context, accountID, eventType string, device dto.DeviceInfo) {
	result := entity.AuditResultSuccess
	if eventType == entity.AuditLoginFailed {
		result = entity.AuditResultFailure
	}
	audit, _ := entity.NewAuditEvent(accountID, accountID, eventType, result, device.IP, device.UA)
	if audit != nil {
		audit.ID = uuid.New().String()
		if err := s.auditRepo.Create(ctx, audit); err != nil {
			slog.Error("failed to write login audit", "error", err)
		}
	}
}

// hashToken 对 token 做 SHA-256 哈希。
func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// maskEmail 对邮箱做脱敏处理。
func maskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return "***"
	}
	local := parts[0]
	if len(local) <= 2 {
		return local[:1] + "***@" + parts[1]
	}
	return local[:2] + "***@" + parts[1]
}
