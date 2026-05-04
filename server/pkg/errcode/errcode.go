package errcode

import (
	"fmt"
	"net/http"
)

// Error 统一错误结构体，实现error接口
type Error struct {
	Code    int    `json:"code"`    // 错误码
	Message string `json:"message"` // 错误信息
}

func (e *Error) Error() string {
	return fmt.Sprintf("错误码: %d, 错误信息: %s", e.Code, e.Message)
}

// NewError 创建新的错误
func NewError(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

// 预定义错误码（遵循阿里错误码规范：五位数字，模块前缀）
var (
	// 成功
	Success = &Error{Code: 20000, Message: "成功"}

	// 客户端错误 4xxxx
	ErrBadRequest    = &Error{Code: 40000, Message: "请求错误"}
	ErrInvalidParam  = &Error{Code: 40001, Message: "参数错误"}
	ErrUnauthorized  = &Error{Code: 40003, Message: "未授权"}
	ErrForbidden     = &Error{Code: 40004, Message: "禁止访问"}
	ErrNotFound      = &Error{Code: 40005, Message: "资源不存在"}
	ErrTooManyReq    = &Error{Code: 40006, Message: "请求过于频繁"}

	// 用户模块错误 401xx
	ErrUserNotFound      = &Error{Code: 40100, Message: "用户不存在"}
	ErrUserAlreadyExists = &Error{Code: 40101, Message: "用户已存在"}
	ErrWrongPassword     = &Error{Code: 40102, Message: "密码错误"}
	ErrEmailAlreadyUsed  = &Error{Code: 40103, Message: "邮箱已被使用"}
	ErrTokenInvalid      = &Error{Code: 40104, Message: "Token无效"}
	ErrTokenExpired      = &Error{Code: 40105, Message: "Token已过期"}

	// 房间模块错误 402xx
	ErrRoomNotFound       = &Error{Code: 40200, Message: "房间不存在"}
	ErrRoomFull           = &Error{Code: 40201, Message: "房间已满"}
	ErrRoomPrivate        = &Error{Code: 40202, Message: "私密房间需要邀请码"}
	ErrInvalidInviteCode  = &Error{Code: 40203, Message: "邀请码无效"}
	ErrAlreadyInRoom      = &Error{Code: 40204, Message: "已在房间中"}
	ErrNotInRoom          = &Error{Code: 40205, Message: "不在房间中"}
	ErrNotRoomHost        = &Error{Code: 40206, Message: "不是房主"}
	ErrRoomClosed         = &Error{Code: 40207, Message: "房间已关闭"}

	// 播放列表错误 403xx
	ErrPlaylistEmpty    = &Error{Code: 40300, Message: "播放列表为空"}
	ErrPlaylistItemNotFound = &Error{Code: 40301, Message: "播放项不存在"}

	// 聊天模块错误 404xx
	ErrMessageTooLong = &Error{Code: 40400, Message: "消息过长"}

	// WebSocket错误 405xx
	ErrWSConnectFailed = &Error{Code: 40500, Message: "WebSocket连接失败"}
	ErrWSNotConnected  = &Error{Code: 40501, Message: "WebSocket未连接"}

	// 服务端错误 5xxxx
	ErrInternalServer = &Error{Code: 50000, Message: "服务器内部错误"}
	ErrDBError        = &Error{Code: 50001, Message: "数据库错误"}
	ErrCacheError     = &Error{Code: 50002, Message: "缓存错误"}
)

// WithMessage 创建带有自定义消息的错误
func (e *Error) WithMessage(message string) *Error {
	return &Error{Code: e.Code, Message: message}
}

// HTTPStatus 返回错误对应的HTTP状态码
func (e *Error) HTTPStatus() int {
	switch {
	case e.Code == 20000:
		return http.StatusOK
	case e.Code >= 40000 && e.Code < 40100:
		return http.StatusBadRequest
	case e.Code >= 40100 && e.Code < 40200:
		return http.StatusUnauthorized
	case e.Code >= 40200 && e.Code < 40300:
		return http.StatusForbidden
	case e.Code >= 40300 && e.Code < 40400:
		return http.StatusNotFound
	case e.Code >= 50000:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}

// Is 判断错误是否为指定错误
func (e *Error) Is(target *Error) bool {
	return e.Code == target.Code
}
