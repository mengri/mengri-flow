package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/repository"
	"mengri-flow/internal/infra/auth"
	"mengri-flow/internal/infra/config"
	"mengri-flow/internal/infra/external/oauth"
	"mengri-flow/pkg/ctxutil"

	"github.com/google/uuid"
)

// AuthServiceImpl 认证应用服务实现。
type authServiceImpl struct {
	accountRepo    repository.AccountRepository         `autowired:""`
	credRepo       repository.CredentialRepository      `autowired:""`
	tokenRepo      repository.ActivationTokenRepository `autowired:""`
	sessionStore   repository.SessionStore              `autowired:""`
	auditRepo      repository.AuditEventRepository      `autowired:""`
	identityRepo   repository.IdentityRepository        `autowired:""`
	otpStore       repository.OTPStore                  `autowired:""`
	smsSender      repository.SMSSender                 `autowired:""`
	oauthProviders oauth.IOAuthProviders                `autowired:""`
	txManager      repository.TransactionManager        `autowired:""`
	jwtManager     auth.IJWTManager                     `autowired:""`
	cfg            *config.AuthConfig                   `autowired:""`
	smsCfg         *config.SMSConfig                    `autowired:""`
	stateStore     repository.OAuthStateStore           `autowired:""`
	bindStore      repository.BindTicketStore           `autowired:""`
}

var _ AuthService = (*authServiceImpl)(nil)

// ValidateActivationToken 验证激活令牌是否有效。
func (s *authServiceImpl) ValidateActivationToken(ctx context.Context, rawToken string) (*dto.ActivationValidateResponse, error) {
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
func (s *authServiceImpl) ConfirmActivation(ctx context.Context, req *dto.ActivationConfirmRequest) (*dto.ActivationConfirmResponse, error) {
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
	ip, ua := ctxutil.ClientIP(ctx), ctxutil.UserAgent(ctx)
	audit, _ := entity.NewAuditEvent("", resp.AccountID, entity.AuditAccountActivated, entity.AuditResultSuccess, ip, ua)
	if audit != nil {
		audit.ID = uuid.New().String()
		if auditErr := s.auditRepo.Create(ctx, audit); auditErr != nil {
			slog.Error("failed to write audit event", "error", auditErr)
		}
	}

	return resp, nil
}

// LoginByPassword 密码登录。
func (s *authServiceImpl) LoginByPassword(ctx context.Context, req *dto.PasswordLoginRequest) (*dto.LoginResponse, error) {
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
func (s *authServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error) {
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
	ip, ua := ctxutil.ClientIP(ctx), ctxutil.UserAgent(ctx)
	audit, _ := entity.NewAuditEvent(accountID, accountID, entity.AuditTokenRefreshed, entity.AuditResultSuccess, ip, ua)
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
func (s *authServiceImpl) Logout(ctx context.Context, accountID, refreshTokenRaw string) error {
	if refreshTokenRaw != "" {
		tokenHash := hashToken(refreshTokenRaw)
		if err := s.sessionStore.RevokeRefreshToken(ctx, tokenHash); err != nil {
			slog.Error("failed to revoke refresh token on logout", "error", err)
		}
	}

	// 审计
	ip, ua := ctxutil.ClientIP(ctx), ctxutil.UserAgent(ctx)
	audit, _ := entity.NewAuditEvent(accountID, accountID, entity.AuditLogout, entity.AuditResultSuccess, ip, ua)
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
func (s *authServiceImpl) issueTokens(ctx context.Context, account *entity.Account, device dto.DeviceInfo) (*dto.LoginResponse, error) {
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

	type deviceMeta struct {
		UA       string `json:"ua"`
		IP       string `json:"ip"`
		DeviceID string `json:"deviceId"`
	}
	deviceJSONBytes, err := json.Marshal(deviceMeta{UA: device.UA, IP: device.IP, DeviceID: device.DeviceID})
	if err != nil {
		return nil, fmt.Errorf("marshal device info: %w", err)
	}
	deviceJSON := string(deviceJSONBytes)

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
func (s *authServiceImpl) recordLoginAudit(ctx context.Context, accountID, eventType string, device dto.DeviceInfo) {
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

// generateOTP 生成指定长度的数字验证码。
func generateOTP(length int) (string, error) {
	const digits = "0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		result[i] = digits[num.Int64()]
	}
	return string(result), nil
}

// hashOTP 对 OTP 做哈希（用于存储）。
func hashOTP(code string) string {
	h := sha256.Sum256([]byte(code))
	return hex.EncodeToString(h[:])
}

// maskPhone 对手机号做脱敏处理。
func maskPhone(phone string) string {
	if len(phone) < 7 {
		return "***"
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// oauthLoginType 将已校验的 provider 名称转换为 LoginType。
// provider 必须已经通过 oauthProviders.GetProvider 校验，此处仅做格式转换。
func oauthLoginType(provider string) entity.LoginType {
	return entity.LoginType(strings.ToLower(provider) + "_oauth")
}

// SendSMSCode 发送短信验证码。
func (s *authServiceImpl) SendSMSCode(ctx context.Context, req *dto.SMSSendRequest) (*dto.SMSSendResponse, error) {
	// 频率限制：60秒内只能发送一次
	count, err := s.otpStore.IncrSendCount(ctx, req.Phone, time.Duration(s.smsCfg.OTPCooldown)*time.Second)
	if err != nil {
		slog.Error("failed to check sms rate limit", "error", err)
		return nil, domainErr.ErrInternal
	}
	if count > 1 {
		return &dto.SMSSendResponse{
			Sent:          false,
			TTLSec:        s.smsCfg.OTPTTL,
			RetryAfterSec: s.smsCfg.OTPCooldown,
		}, domainErr.ErrOTPTooFrequent
	}

	// 生成验证码
	code, err := generateOTP(s.smsCfg.OTPLength)
	if err != nil {
		slog.Error("failed to generate otp", "error", err)
		return nil, domainErr.ErrInternal
	}

	// 存储到 Redis（哈希后存储）
	codeHash := hashOTP(code)
	if err := s.otpStore.Save(ctx, req.Scene, req.Phone, codeHash, time.Duration(s.smsCfg.OTPTTL)*time.Second); err != nil {
		slog.Error("failed to save otp", "error", err)
		return nil, domainErr.ErrInternal
	}

	// 发送短信
	if err := s.smsSender.SendOTP(ctx, req.Phone, code); err != nil {
		slog.Error("failed to send sms", "phone", req.Phone, "error", err)
		return nil, domainErr.ErrInternal
	}

	return &dto.SMSSendResponse{
		Sent:          true,
		TTLSec:        s.smsCfg.OTPTTL,
		RetryAfterSec: s.smsCfg.OTPCooldown,
	}, nil
}

// LoginBySMS 短信验证码登录。
func (s *authServiceImpl) LoginBySMS(ctx context.Context, req *dto.SMSLoginRequest) (*dto.LoginResponse, error) {
	// 1. 验证验证码
	storedHash, err := s.otpStore.Get(ctx, "login", req.Phone)
	if err != nil {
		slog.Error("failed to get otp", "error", err)
		return nil, domainErr.ErrOTPInvalid
	}
	if storedHash == "" {
		return nil, domainErr.ErrOTPExpired
	}

	// 验证并删除验证码（一次性）
	inputHash := hashOTP(req.Code)
	if inputHash != storedHash {
		// 记录审计
		s.recordLoginAudit(ctx, "", entity.AuditLoginFailed, req.DeviceInfo)
		return nil, domainErr.ErrOTPInvalid
	}
	// 删除已使用的验证码
	s.otpStore.Delete(ctx, "login", req.Phone)

	// 2. 查找手机号对应的身份
	identity, err := s.identityRepo.GetByProviderID(ctx, entity.LoginTypeSMS, req.Phone)
	if err != nil {
		// 手机号未绑定任何账号
		return nil, domainErr.ErrIdentityNotBound
	}

	// 3. 查询账号
	account, err := s.accountRepo.GetByID(ctx, identity.AccountID)
	if err != nil {
		return nil, domainErr.ErrAccountNotFound
	}

	// 4. 检查账号状态
	if !account.CanLogin() {
		if account.Status == entity.AccountStatusPendingActivation {
			return nil, domainErr.ErrAccountNotActivated
		}
		if account.Status == entity.AccountStatusLocked {
			return nil, domainErr.ErrAccountLocked
		}
		if account.Status == entity.AccountStatusDisabled {
			return nil, domainErr.ErrAccountDisabled
		}
		return nil, domainErr.ErrAccountLocked
	}

	// 5. 签发 Token
	loginResp, err := s.issueTokens(ctx, account, req.DeviceInfo)
	if err != nil {
		return nil, err
	}

	// 6. 记录登录成功审计
	s.recordLoginAudit(ctx, account.ID, entity.AuditLoginSuccess, req.DeviceInfo)

	return loginResp, nil
}

// GetOAuthURL 获取第三方授权地址。
func (s *authServiceImpl) GetOAuthURL(ctx context.Context, provider, scene, redirectURI string) (*dto.OAuthURLResponse, error) {
	oauthProvider, ok := s.oauthProviders.GetProvider(provider)
	if !ok {
		return nil, domainErr.ErrOAuthProviderNotSupported
	}

	// 生成并存储 state（CSRF 防护）
	state, err := s.stateStore.Generate(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate oauth state: %w", err)
	}

	authURL := oauthProvider.GetAuthURL(state, redirectURI)

	return &dto.OAuthURLResponse{
		AuthURL:  authURL,
		State:    state,
		ExpireAt: time.Now().Add(5 * time.Minute).Format(time.RFC3339),
	}, nil
}

// HandleOAuthCallback 处理第三方回调。
func (s *authServiceImpl) HandleOAuthCallback(ctx context.Context, provider, code, state string) (*dto.OAuthCallbackResponse, error) {
	// 1. 验证 state
	if err := s.stateStore.Validate(ctx, state); err != nil {
		return nil, domainErr.ErrOAuthStateInvalid
	}

	oauthProvider, ok := s.oauthProviders.GetProvider(provider)
	if !ok {
		return nil, domainErr.ErrOAuthProviderNotSupported
	}

	// 2. 换取用户信息
	userInfo, err := oauthProvider.ExchangeCode(ctx, code)
	if err != nil {
		slog.Error("failed to exchange oauth code", "provider", provider, "error", err)
		return nil, domainErr.ErrOAuthExchangeFailed
	}

	// 3. 查找是否已绑定（provider 已通过 GetProvider 校验，使用规范化函数转换避免直接拼接）
	identity, err := s.identityRepo.GetByProviderID(ctx, oauthLoginType(provider), userInfo.ProviderUserID)
	if err != nil {

		bindTicket, err := s.bindStore.Generate(ctx, &repository.BindTicketData{
			Provider:   provider,
			ExternalID: userInfo.ProviderUserID,
			Nickname:   userInfo.DisplayName,
			AvatarURL:  userInfo.AvatarURL,
		})
		if err != nil {
			return nil, fmt.Errorf("generate bind ticket: %w", err)
		}

		return &dto.OAuthCallbackResponse{
			Result:     "NEED_BIND_EXISTING_ACCOUNT",
			Provider:   provider,
			BindTicket: bindTicket,
			ExpireAt:   time.Now().Add(5 * time.Minute).Format(time.RFC3339),
		}, nil
	}

	// 4. 已绑定，查询账号
	account, err := s.accountRepo.GetByID(ctx, identity.AccountID)
	if err != nil {
		return nil, domainErr.ErrAccountNotFound
	}

	// 5. 检查账号状态
	if !account.CanLogin() {
		if account.Status == entity.AccountStatusPendingActivation {
			return nil, domainErr.ErrAccountNotActivated
		}
		if account.Status == entity.AccountStatusLocked {
			return nil, domainErr.ErrAccountLocked
		}
		if account.Status == entity.AccountStatusDisabled {
			return nil, domainErr.ErrAccountDisabled
		}
		return nil, domainErr.ErrAccountLocked
	}

	// 6. 签发 Token
	ip, ua := ctxutil.ClientIP(ctx), ctxutil.UserAgent(ctx)
	loginResp, err := s.issueTokens(ctx, account, dto.DeviceInfo{IP: ip, UA: ua})
	if err != nil {
		return nil, err
	}

	// 7. 记录登录审计
	s.recordLoginAudit(ctx, account.ID, entity.AuditLoginSuccess, dto.DeviceInfo{IP: ip, UA: ua})

	return &dto.OAuthCallbackResponse{
		Result:       "LOGIN_SUCCESS",
		AccessToken:  loginResp.AccessToken,
		RefreshToken: loginResp.RefreshToken,
		ExpiresIn:    loginResp.ExpiresIn,
		Account:      &loginResp.Account,
	}, nil
}
