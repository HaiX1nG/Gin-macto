package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/pkg/errcode"
)

// Response 统一响应结构体
type Response struct {
	Code    int         `json:"code"`    // 错误码
	Message string      `json:"message"` // 错误信息
	Data    interface{} `json:"data"`    // 响应数据
}

// PageData 分页数据结构
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.Success.Code,
		Message: errcode.Success.Message,
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.Success.Code,
		Message: message,
		Data:    data,
	})
}

// Fail 失败响应
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
func FailWithData(c *gin.Context, err error, data interface{}) {
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
func Page(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	Success(c, PageData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}