package handler

import (
	"errors"
	"strconv"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/app/service"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/pkg/response"

	"github.com/gin-gonic/gin"
)

// MeHandlerImpl 账号中心 HTTP 处理器实现。
type MeHandlerImpl struct {
	meService service.IMeService `autowired:""`
}

var _ IMeHandler = (*MeHandlerImpl)(nil)

// GetProfile 获取当前用户资料。
// @Summary 获取资料
// @Tags Me
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=dto.ProfileResponse}
// @Router /api/v1/me/profile [get]
func (h *MeHandlerImpl) GetProfile(c *gin.Context) {
	accountID := getAccountID(c)
	resp, err := h.meService.GetProfile(c.Request.Context(), accountID)
	if err != nil {
		handleMeDomainError(c, err)
		return
	}
	response.OK(c, resp)
}

// ListIdentities 获取当前用户登录身份列表。
// @Summary 身份列表
// @Tags Me
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=dto.IdentityListResponse}
// @Router /api/v1/me/identities [get]
func (h *MeHandlerImpl) ListIdentities(c *gin.Context) {
	accountID := getAccountID(c)
	resp, err := h.meService.ListIdentities(c.Request.Context(), accountID)
	if err != nil {
		handleMeDomainError(c, err)
		return
	}
	response.OK(c, resp)
}

// ChangePassword 修改密码。
// @Summary 修改密码
// @Tags Me
// @Accept json
// @Produce json
// @Param request body dto.ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} response.Response{data=dto.ChangePasswordResponse}
// @Router /api/v1/me/password/change [post]
func (h *MeHandlerImpl) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	accountID := getAccountID(c)
	resp, err := h.meService.ChangePassword(c.Request.Context(), accountID, &req)
	if err != nil {
		handleMeDomainError(c, err)
		return
	}
	response.OK(c, resp)
}

// SecurityVerify 二次安全验证。
// @Summary 安全验证
// @Tags Me
// @Accept json
// @Produce json
// @Param request body dto.SecurityVerifyRequest true "验证请求"
// @Success 200 {object} response.Response{data=dto.SecurityTicketResponse}
// @Router /api/v1/me/security/verify [post]
func (h *MeHandlerImpl) SecurityVerify(c *gin.Context) {
	var req dto.SecurityVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	accountID := getAccountID(c)
	resp, err := h.meService.SecurityVerify(c.Request.Context(), accountID, req.Password)
	if err != nil {
		handleMeDomainError(c, err)
		return
	}
	response.OK(c, resp)
}

// LoginHistory 查询登录记录。
// @Summary 登录历史
// @Tags Me
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=dto.AuditEventListResponse}
// @Router /api/v1/me/security/logins [get]
func (h *MeHandlerImpl) LoginHistory(c *gin.Context) {
	accountID := getAccountID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	resp, err := h.meService.LoginHistory(c.Request.Context(), accountID, page, pageSize)
	if err != nil {
		handleMeDomainError(c, err)
		return
	}
	response.OK(c, resp)
}

// handleMeDomainError 将领域错误映射到 HTTP 响应。
func handleMeDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainErr.ErrAccountNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, domainErr.ErrCredentialsInvalid):
		response.Unauthorized(c, err.Error())
	case errors.Is(err, domainErr.ErrCannotUnbindLast):
		response.BadRequest(c, err.Error())
	case errors.Is(err, domainErr.ErrSecurityTicketInvalid):
		response.BadRequest(c, err.Error())
	case errors.Is(err, domainErr.ErrInvalidPassword):
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, "internal server error")
	}
}
