package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应格式 { code, data, msg }
type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

// OK 成功响应
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Data: data,
		Msg:  "success",
	})
}

// OKWithMsg 成功响应并附带消息
func OKWithMsg(c *gin.Context, data interface{}, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Data: data,
		Msg:  msg,
	})
}

// Fail 失败响应
func Fail(c *gin.Context, httpCode int, bizCode int, msg string) {
	c.JSON(httpCode, Response{
		Code: bizCode,
		Data: nil,
		Msg:  msg,
	})
}

// BadRequest 400 错误
func BadRequest(c *gin.Context, msg string) {
	Fail(c, http.StatusBadRequest, 400, msg)
}

// NotFound 404 错误
func NotFound(c *gin.Context, msg string) {
	Fail(c, http.StatusNotFound, 404, msg)
}

// Unauthorized 401 错误
func Unauthorized(c *gin.Context, msg string) {
	Fail(c, http.StatusUnauthorized, 401, msg)
}

// Forbidden 403 错误
func Forbidden(c *gin.Context, msg string) {
	Fail(c, http.StatusForbidden, 403, msg)
}

// InternalError 500 错误
func InternalError(c *gin.Context, msg string) {
	Fail(c, http.StatusInternalServerError, 500, msg)
}

// Conflict 409 错误
func Conflict(c *gin.Context, msg string) {
	Fail(c, http.StatusConflict, 409, msg)
}
