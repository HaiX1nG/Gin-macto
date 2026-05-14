package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/pkg/errcode"
)

// Response 统一响应结构体
// 所有API响应都使用此结构体进行封装，确保响应格式一致
type Response struct {
	Code    int    `json:"code"`    // Code 错误码，20000表示成功，其他值表示各类错误
	Message string `json:"message"` // Message 响应消息，描述操作结果或错误原因
	Data    any    `json:"data"`    // Data 响应数据，成功时返回业务数据，失败时为nil
}

// PageData 分页数据结构
// 用于分页查询接口的响应数据封装
type PageData struct {
	List     any   `json:"list"`     // List 当前页数据列表
	Total    int64 `json:"total"`    // Total 符合条件的总记录数
	Page     int   `json:"page"`     // Page 当前页码，从1开始
	PageSize int   `json:"pageSize"` // PageSize 每页记录数
}

// Success 成功响应
// 返回HTTP 200状态码和成功响应结构体
// 参数：
//   - c: Gin上下文
//   - data: 响应数据，可以是任意类型
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.Success.Code,
		Message: errcode.Success.Message,
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（自定义消息）
// 返回HTTP 200状态码和成功响应结构体，使用自定义消息
// 参数：
//   - c: Gin上下文
//   - message: 自定义成功消息
//   - data: 响应数据，可以是任意类型
func SuccessWithMessage(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.Success.Code,
		Message: message,
		Data:    data,
	})
}

// Fail 失败响应
// 根据错误类型返回对应的HTTP状态码和错误信息
// 参数：
//   - c: Gin上下文
//   - err: 错误对象，如果是errcode.Error类型则使用其Code和HTTPStatus，否则返回500错误
func Fail(c *gin.Context, err error) {
	var e *errcode.Error
	if errors.As(err, &e) {
		c.JSON(e.HTTPStatus(), Response{
			Code:    e.Code,
			Message: e.Message,
			Data:    nil,
		})
		return
	}
	// 默认返回内部服务器错误
	c.JSON(http.StatusInternalServerError, Response{
		Code:    errcode.ErrInternalServer.Code,
		Message: err.Error(),
		Data:    nil,
	})
}

// FailWithData 失败响应（带数据）
// 根据错误类型返回对应的HTTP状态码和错误信息，同时携带附加数据
// 参数：
//   - c: Gin上下文
//   - err: 错误对象，如果是errcode.Error类型则使用其Code和HTTPStatus，否则返回500错误
//   - data: 附加数据，用于返回错误详情或调试信息
func FailWithData(c *gin.Context, err error, data any) {
	var e *errcode.Error
	if errors.As(err, &e) {
		c.JSON(e.HTTPStatus(), Response{
			Code:    e.Code,
			Message: e.Message,
			Data:    data,
		})
		return
	}
	c.JSON(http.StatusInternalServerError, Response{
		Code:    errcode.ErrInternalServer.Code,
		Message: err.Error(),
		Data:    data,
	})
}

// FailWithMessage 失败响应（自定义消息）
// 根据错误类型返回对应的HTTP状态码，使用自定义消息覆盖原错误消息
// 参数：
//   - c: Gin上下文
//   - err: 错误对象，用于获取HTTP状态码和错误码
//   - message: 自定义错误消息
func FailWithMessage(c *gin.Context, err error, message string) {
	var e *errcode.Error
	if errors.As(err, &e) {
		c.JSON(e.HTTPStatus(), Response{
			Code:    e.Code,
			Message: message,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusInternalServerError, Response{
		Code:    errcode.ErrInternalServer.Code,
		Message: message,
		Data:    nil,
	})
}

// Page 分页响应
// 返回分页数据结构的成功响应
// 参数：
//   - c: Gin上下文
//   - list: 当前页数据列表
//   - total: 符合条件的总记录数
//   - page: 当前页码，从1开始
//   - pageSize: 每页记录数
func Page(c *gin.Context, list any, total int64, page, pageSize int) {
	Success(c, PageData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
