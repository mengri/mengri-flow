package handler

import (
	"errors"
	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/app/service"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandlerImpl HTTP 处理器实现，属于 Ports 层，负责参数绑定和响应转换。
type UserHandlerImpl struct {
	userService service.UserService `autowired:""`
}

// 编译期接口合规检查
var _ UserHandler = (*UserHandlerImpl)(nil)

// Create 创建用户
// @Summary 创建用户
// @Tags User
// @Accept json
// @Produce json
// @Param body body dto.CreateUserRequest true "创建用户请求"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Router /api/v1/users [post]
func (h *UserHandlerImpl) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.OK(c, result)
}

// GetByID 根据 ID 获取用户
// @Summary 获取用户详情
// @Tags User
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Router /api/v1/users/{id} [get]
func (h *UserHandlerImpl) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	result, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.OK(c, result)
}

// List 获取用户列表
// @Summary 获取用户列表
// @Tags User
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Response{data=dto.ListUsersResponse}
// @Router /api/v1/users [get]
func (h *UserHandlerImpl) List(c *gin.Context) {
	var req dto.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.userService.ListUsers(c.Request.Context(), &req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.OK(c, result)
}

// Update 更新用户
// @Summary 更新用户
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param body body dto.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Router /api/v1/users/{id} [put]
func (h *UserHandlerImpl) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.userService.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.OK(c, result)
}

// Delete 删除用户
// @Summary 删除用户
// @Tags User
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id} [delete]
func (h *UserHandlerImpl) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), id); err != nil {
		handleDomainError(c, err)
		return
	}

	response.OKWithMsg(c, nil, "user deleted")
}

// handleDomainError 将领域错误映射为 HTTP 响应
func handleDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainErr.ErrUserNotFound), errors.Is(err, domainErr.ErrNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, domainErr.ErrEmailTaken), errors.Is(err, domainErr.ErrConflict):
		response.Conflict(c, err.Error())
	case errors.Is(err, domainErr.ErrInvalidUsername),
		errors.Is(err, domainErr.ErrInvalidEmail),
		errors.Is(err, domainErr.ErrInvalidPassword):
		response.BadRequest(c, err.Error())
	case errors.Is(err, domainErr.ErrUnauthorized):
		response.Unauthorized(c, err.Error())
	case errors.Is(err, domainErr.ErrForbidden), errors.Is(err, domainErr.ErrUserBanned):
		response.Forbidden(c, err.Error())
	default:
		response.InternalError(c, "internal server error")
	}
}
