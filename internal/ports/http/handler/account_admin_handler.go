package handler

import (
	"errors"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/app/service"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/pkg/response"

	"github.com/gin-gonic/gin"
)

// AccountAdminHandlerImpl 管理员账号管理 HTTP 处理器实现。
type AccountAdminHandlerImpl struct {
	adminService service.IAccountAdminService `autowired:""`
}

var _ IAccountAdminHandler = (*AccountAdminHandlerImpl)(nil)

// Create 管理员创建账号。
// @Summary 创建账号
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body dto.CreateAccountRequest true "创建账号请求"
// @Success 200 {object} response.Response{data=dto.AccountResponse}
// @Router /api/v1/admin/accounts [post]
func (h *AccountAdminHandlerImpl) Create(c *gin.Context) {
	var req dto.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	operatorID := getAccountID(c)
	resp, err := h.adminService.CreateAccount(c.Request.Context(), &req, operatorID)
	if err != nil {
		handleAdminDomainError(c, err)
		return
	}
	response.OK(c, resp)
}

// List 查询账号列表。
// @Summary 账号列表
// @Tags Admin
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param status query string false "状态过滤"
// @Param keyword query string false "关键字"
// @Success 200 {object} response.Response{data=dto.ListAccountsResponse}
// @Router /api/v1/admin/accounts [get]
func (h *AccountAdminHandlerImpl) List(c *gin.Context) {
	var req dto.ListAccountsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.adminService.ListAccounts(c.Request.Context(), &req)
	if err != nil {
		handleAdminDomainError(c, err)
		return
	}
	response.OK(c, resp)
}

// GetDetail 获取账号详情。
// @Summary 账号详情
// @Tags Admin
// @Accept json
// @Produce json
// @Param accountId path string true "账号 ID"
// @Success 200 {object} response.Response{data=dto.AccountDetailResponse}
// @Router /api/v1/admin/accounts/{accountId} [get]
func (h *AccountAdminHandlerImpl) GetDetail(c *gin.Context) {
	accountID := c.Param("accountId")
	if accountID == "" {
		response.BadRequest(c, "accountId is required")
		return
	}

	resp, err := h.adminService.GetAccountDetail(c.Request.Context(), accountID)
	if err != nil {
		handleAdminDomainError(c, err)
		return
	}
	response.OK(c, resp)
}

// ChangeStatus 变更账号状态。
// @Summary 变更账号状态
// @Tags Admin
// @Accept json
// @Produce json
// @Param accountId path string true "账号 ID"
// @Param request body dto.ChangeStatusRequest true "变更状态请求"
// @Success 200 {object} response.Response{data=dto.AccountResponse}
// @Router /api/v1/admin/accounts/{accountId}/status [put]
func (h *AccountAdminHandlerImpl) ChangeStatus(c *gin.Context) {
	accountID := c.Param("accountId")
	if accountID == "" {
		response.BadRequest(c, "accountId is required")
		return
	}

	var req dto.ChangeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	operatorID := getAccountID(c)
	resp, err := h.adminService.ChangeAccountStatus(c.Request.Context(), accountID, &req, operatorID)
	if err != nil {
		handleAdminDomainError(c, err)
		return
	}
	response.OK(c, resp)
}

// ResendActivation 重发激活邮件。
// @Summary 重发激活邮件
// @Tags Admin
// @Accept json
// @Produce json
// @Param accountId path string true "账号 ID"
// @Param request body dto.ResendActivationRequest true "重发请求"
// @Success 200 {object} response.Response{data=dto.ResendActivationResponse}
// @Router /api/v1/admin/accounts/{accountId}/activation/resend [post]
func (h *AccountAdminHandlerImpl) ResendActivation(c *gin.Context) {
	accountID := c.Param("accountId")
	if accountID == "" {
		response.BadRequest(c, "accountId is required")
		return
	}

	var req dto.ResendActivationRequest
	_ = c.ShouldBindJSON(&req) // reason is optional

	operatorID := getAccountID(c)
	resp, err := h.adminService.ResendActivation(c.Request.Context(), accountID, req.Reason, operatorID)
	if err != nil {
		handleAdminDomainError(c, err)
		return
	}
	response.OK(c, resp)
}

// ListAuditEvents 查询审计事件列表。
// @Summary 审计事件列表
// @Tags Admin
// @Accept json
// @Produce json
// @Param accountId query string false "账号 ID"
// @Param eventType query string false "事件类型"
// @Param from query string false "开始时间"
// @Param to query string false "结束时间"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=dto.AuditEventListResponse}
// @Router /api/v1/admin/audit/events [get]
func (h *AccountAdminHandlerImpl) ListAuditEvents(c *gin.Context) {
	var req dto.AuditEventFilter
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.adminService.ListAuditEvents(c.Request.Context(), &req)
	if err != nil {
		handleAdminDomainError(c, err)
		return
	}
	response.OK(c, resp)
}

// getAccountID 从 gin.Context 获取当前登录用户 ID。
func getAccountID(c *gin.Context) string {
	val, _ := c.Get("accountID")
	id, _ := val.(string)
	return id
}

// handleAdminDomainError 将领域错误映射到 HTTP 响应。
func handleAdminDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainErr.ErrAccountNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, domainErr.ErrEmailTaken):
		response.Conflict(c, err.Error())
	case errors.Is(err, domainErr.ErrUsernameTaken):
		response.Conflict(c, err.Error())
	case errors.Is(err, domainErr.ErrConflict):
		response.Conflict(c, err.Error())
	case errors.Is(err, domainErr.ErrAlreadyActivated):
		response.Conflict(c, err.Error())
	case errors.Is(err, domainErr.ErrInvalidStatusTransition):
		response.BadRequest(c, err.Error())
	case errors.Is(err, domainErr.ErrActivationTooFrequent):
		response.BadRequest(c, err.Error())
	case errors.Is(err, domainErr.ErrInvalidUsername):
		response.BadRequest(c, err.Error())
	case errors.Is(err, domainErr.ErrInvalidDisplayName):
		response.BadRequest(c, err.Error())
	case errors.Is(err, domainErr.ErrInvalidEmail):
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, "internal server error")
	}
}
