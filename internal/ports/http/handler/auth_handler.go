package handler

import (
	"errors"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/app/service"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthHandlerImpl 认证 HTTP 处理器实现。
type AuthHandlerImpl struct {
	authService service.AuthService `autowired:""`
}

var _ AuthHandler = (*AuthHandlerImpl)(nil)

// ValidateActivation 激活链接预校验。
// @Summary 校验激活令牌
// @Tags Auth
// @Accept json
// @Produce json
// @Param token query string true "激活令牌"
// @Success 200 {object} response.Response{data=dto.ActivationValidateResponse}
// @Router /api/v1/auth/activation/validate [get]
func (h *AuthHandlerImpl) ValidateActivation(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.BadRequest(c, "token is required")
		return
	}

	resp, err := h.authService.ValidateActivationToken(c.Request.Context(), token)
	if err != nil {
		handleAuthDomainError(c, err)
		return
	}
	response.OK(c, resp)
}

// ConfirmActivation 确认激活并设置密码。
// @Summary 确认激活
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ActivationConfirmRequest true "激活请求"
// @Success 200 {object} response.Response{data=dto.ActivationConfirmResponse}
// @Router /api/v1/auth/activation/confirm [post]
func (h *AuthHandlerImpl) ConfirmActivation(c *gin.Context) {
	var req dto.ActivationConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authService.ConfirmActivation(c.Request.Context(), &req)
	if err != nil {
		handleAuthDomainError(c, err)
		return
	}
	response.OK(c, resp)
}

// LoginByPassword 密码登录。
// @Summary 密码登录
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.PasswordLoginRequest true "登录请求"
// @Success 200 {object} response.Response{data=dto.LoginResponse}
// @Router /api/v1/auth/login/password [post]
func (h *AuthHandlerImpl) LoginByPassword(c *gin.Context) {
	var req dto.PasswordLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 自动填充 IP
	if req.DeviceInfo.IP == "" {
		req.DeviceInfo.IP = c.ClientIP()
	}
	if req.DeviceInfo.UA == "" {
		req.DeviceInfo.UA = c.GetHeader("User-Agent")
	}

	resp, err := h.authService.LoginByPassword(c.Request.Context(), &req)
	if err != nil {
		handleAuthDomainError(c, err)
		return
	}
	response.OK(c, resp)
}

// RefreshToken 刷新 Token。
// @Summary 刷新 Token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "刷新请求"
// @Success 200 {object} response.Response{data=dto.LoginResponse}
// @Router /api/v1/auth/token/refresh [post]
func (h *AuthHandlerImpl) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		handleAuthDomainError(c, err)
		return
	}
	response.OK(c, resp)
}

// Logout 登出。
// @Summary 登出
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/auth/logout [post]
func (h *AuthHandlerImpl) Logout(c *gin.Context) {
	accountID, _ := c.Get("accountId")
	accID, _ := accountID.(string)

	// 尝试从请求体获取 refreshToken
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	_ = c.ShouldBindJSON(&body)

	if err := h.authService.Logout(c.Request.Context(), accID, body.RefreshToken); err != nil {
		handleAuthDomainError(c, err)
		return
	}
	response.OK(c, gin.H{"success": true})
}

// handleAuthDomainError 将领域错误映射到 HTTP 响应。
func handleAuthDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainErr.ErrCredentialsInvalid):
		response.Unauthorized(c, err.Error())
	case errors.Is(err, domainErr.ErrAccountNotActivated):
		response.Forbidden(c, err.Error())
	case errors.Is(err, domainErr.ErrAccountLocked):
		response.Forbidden(c, err.Error())
	case errors.Is(err, domainErr.ErrAccountDisabled):
		response.Forbidden(c, err.Error())
	case errors.Is(err, domainErr.ErrSessionExpired):
		response.Unauthorized(c, err.Error())
	case errors.Is(err, domainErr.ErrActivationTokenInvalid):
		response.BadRequest(c, err.Error())
	case errors.Is(err, domainErr.ErrActivationTokenExpired):
		response.BadRequest(c, err.Error())
	case errors.Is(err, domainErr.ErrActivationTokenUsed):
		response.BadRequest(c, err.Error())
	case errors.Is(err, domainErr.ErrAlreadyActivated):
		response.Conflict(c, err.Error())
	case errors.Is(err, domainErr.ErrAccountNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, domainErr.ErrInvalidPassword):
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, "internal server error")
	}
}
