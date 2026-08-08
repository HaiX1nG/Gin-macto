package errcode

import (
	"fmt"
	"net/http"
)

// Error 统一错误结构体，实现error接口
// 用于API响应中的错误信息封装，遵循阿里巴巴错误码规范
type Error struct {
	Code    int    `json:"code"`    // Code 错误码，五位数字，遵循阿里错误码规范
	Message string `json:"message"` // Message 错误信息，描述错误详情
}

// Error 实现error接口，返回格式化的错误字符串
func (e *Error) Error() string {
	return fmt.Sprintf("错误码: %d, 错误信息: %s", e.Code, e.Message)
}

// NewError 创建新的错误实例
// 参数：
//   - code: 错误码，五位数字
//   - message: 错误信息
//
// 返回：Error指针
func NewError(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

// 预定义错误码（遵循阿里错误码规范：五位数字，模块前缀）
var (
	// Success 成功响应，错误码20000
	Success = &Error{Code: 20000, Message: "成功"}

	// 客户端错误 4xxxx
	// ErrBadRequest 请求错误，客户端发送了无效的请求
	ErrBadRequest = &Error{Code: 40000, Message: "请求错误"}
	// ErrInvalidParam 参数错误，请求参数校验失败
	ErrInvalidParam = &Error{Code: 40001, Message: "参数错误"}
	// ErrUnauthorized 未授权，缺少有效的身份认证信息
	ErrUnauthorized = &Error{Code: 40003, Message: "未授权"}
	// ErrForbidden 禁止访问，无权限访问该资源
	ErrForbidden = &Error{Code: 40004, Message: "禁止访问"}
	// ErrNotFound 资源不存在，请求的资源未找到
	ErrNotFound = &Error{Code: 40005, Message: "资源不存在"}
	// ErrTooManyReq 请求过于频繁，触发限流
	ErrTooManyReq = &Error{Code: 40006, Message: "请求过于频繁"}

	// 用户模块错误 401xx
	// ErrUserNotFound 用户不存在，查询的用户ID在数据库中未找到
	ErrUserNotFound = &Error{Code: 40100, Message: "用户不存在"}
	// ErrUserAlreadyExists 用户已存在，注册时用户名或邮箱已被使用
	ErrUserAlreadyExists = &Error{Code: 40101, Message: "用户已存在"}
	// ErrWrongPassword 密码错误，登录时密码校验失败
	ErrWrongPassword = &Error{Code: 40102, Message: "密码错误"}
	// ErrEmailAlreadyUsed 邮箱已被使用，该邮箱已注册其他账户
	ErrEmailAlreadyUsed = &Error{Code: 40103, Message: "邮箱已被使用"}
	// ErrTokenInvalid Token无效，JWT解析失败或格式错误
	ErrTokenInvalid = &Error{Code: 40104, Message: "Token无效"}
	// ErrTokenExpired Token已过期，JWT已超过有效期
	ErrTokenExpired = &Error{Code: 40105, Message: "Token已过期"}
	// ErrTokenBlacklisted Token已被加入黑名单，已退出登录或被强制下线
	ErrTokenBlacklisted = &Error{Code: 40106, Message: "Token已失效"}

	// 房间模块错误 402xx
	// ErrRoomNotFound 房间不存在，查询的房间ID在数据库中未找到
	ErrRoomNotFound = &Error{Code: 40200, Message: "房间不存在"}
	// ErrRoomFull 房间已满，房间当前参与者数量已达到上限
	ErrRoomFull = &Error{Code: 40201, Message: "房间已满"}
	// ErrRoomPrivate 私密房间需要邀请码，访问私密房间时未提供邀请码
	ErrRoomPrivate = &Error{Code: 40202, Message: "私密房间需要邀请码"}
	// ErrInvalidInviteCode 邀请码无效，提供的邀请码不存在或已过期
	ErrInvalidInviteCode = &Error{Code: 40203, Message: "邀请码无效"}
	// ErrAlreadyInRoom 已在房间中，用户尝试加入已参与的房间
	ErrAlreadyInRoom = &Error{Code: 40204, Message: "已在房间中"}
	// ErrNotInRoom 不在房间中，用户尝试退出未参与的房间
	ErrNotInRoom = &Error{Code: 40205, Message: "不在房间中"}
	// ErrNotRoomHost 不是房主，非房主用户尝试执行房主专属操作
	ErrNotRoomHost = &Error{Code: 40206, Message: "不是房主"}
	// ErrRoomClosed 房间已关闭，尝试操作已关闭的房间
	ErrRoomClosed = &Error{Code: 40207, Message: "房间已关闭"}

	// 播放列表错误 403xx
	// ErrPlaylistEmpty 播放列表为空，房间中没有待播放的音乐
	ErrPlaylistEmpty = &Error{Code: 40300, Message: "播放列表为空"}
	// ErrPlaylistItemNotFound 播放项不存在，指定的播放项ID在数据库中未找到
	ErrPlaylistItemNotFound = &Error{Code: 40301, Message: "播放项不存在"}

	// 聊天模块错误 404xx
	// ErrMessageTooLong 消息过长，聊天消息超过最大允许长度
	ErrMessageTooLong = &Error{Code: 40400, Message: "消息过长"}

	// 好友模块错误 406xx
	// ErrFriendNotFound 好友不存在，指定的好友关系不存在
	ErrFriendNotFound = &Error{Code: 40600, Message: "好友不存在"}
	// ErrFriendRequestExists 好友请求已存在，双方之间已有待处理的好友请求
	ErrFriendRequestExists = &Error{Code: 40601, Message: "好友请求已存在"}
	// ErrFriendRequestHandled 好友请求已处理，该好友请求已被接受或拒绝
	ErrFriendRequestHandled = &Error{Code: 40602, Message: "好友请求已处理"}
	// ErrNotFriend 不是好友关系，尝试对非好友用户执行好友专属操作
	ErrNotFriend = &Error{Code: 40603, Message: "不是好友关系"}

	// 服务器模块错误 407xx
	// ErrServerNotFound 服务器不存在，查询的服务器ID在数据库中未找到
	ErrServerNotFound = &Error{Code: 40700, Message: "服务器不存在"}
	// ErrServerFull 服务器已满，服务器当前成员数量已达到上限
	ErrServerFull = &Error{Code: 40701, Message: "服务器已满"}
	// ErrInvalidServerInviteCode 服务器邀请码无效
	ErrInvalidServerInviteCode = &Error{Code: 40702, Message: "邀请码无效"}
	// ErrAlreadyInServer 已在服务器中，用户尝试加入已参与的服务器
	ErrAlreadyInServer = &Error{Code: 40703, Message: "已在服务器中"}
	// ErrNotInServer 不在服务器中，用户尝试退出未参与的服务器
	ErrNotInServer = &Error{Code: 40704, Message: "不在服务器中"}
	// ErrNotServerOwner 不是服务器拥有者，非 owner 用户尝试执行 owner 专属操作
	ErrNotServerOwner = &Error{Code: 40705, Message: "不是服务器拥有者"}
	// ErrServerMemberNotFound 服务器成员不存在
	ErrServerMemberNotFound = &Error{Code: 40706, Message: "服务器成员不存在"}
	// ErrRoleNotFound 角色不存在
	ErrRoleNotFound = &Error{Code: 40707, Message: "角色不存在"}
	// ErrCannotDeleteDefaultRole 不能删除默认角色
	ErrCannotDeleteDefaultRole = &Error{Code: 40708, Message: "不能删除默认角色"}
	// ErrChannelNotFound 频道不存在
	ErrChannelNotFound = &Error{Code: 40709, Message: "频道不存在"}
	// ErrMessageNotFound 消息不存在
	ErrMessageNotFound = &Error{Code: 40710, Message: "消息不存在"}
	// ErrNoPermission 无权限执行此操作
	ErrNoPermission = &Error{Code: 40711, Message: "无权限"}

	// WebSocket错误 405xx
	// ErrWSConnectFailed WebSocket连接失败，建立WebSocket连接时发生错误
	ErrWSConnectFailed = &Error{Code: 40500, Message: "WebSocket连接失败"}
	// ErrWSNotConnected WebSocket未连接，尝试在未建立连接时发送消息
	ErrWSNotConnected = &Error{Code: 40501, Message: "WebSocket未连接"}

	// 服务端错误 5xxxx
	// ErrInternalServer 服务器内部错误，未知的服务器异常
	ErrInternalServer = &Error{Code: 50000, Message: "服务器内部错误"}
	// ErrDBError 数据库错误，数据库操作失败
	ErrDBError = &Error{Code: 50001, Message: "数据库错误"}
	// ErrCacheError 缓存错误，Redis操作失败
	ErrCacheError = &Error{Code: 50002, Message: "缓存错误"}
)

// WithMessage 创建带有自定义消息的错误副本
// 参数：
//   - message: 自定义错误信息
//
// 返回：保留原错误码但使用新消息的Error指针
func (e *Error) WithMessage(message string) *Error {
	return &Error{Code: e.Code, Message: message}
}

// HTTPStatus 返回错误对应的HTTP状态码
// 映射规则：
//   - 特定错误码精确映射（如 40003 -> 401, 40004 -> 403）
//   - 通用客户端错误返回 400
//   - 服务端错误返回 500
func (e *Error) HTTPStatus() int {
	// 特定错误码精确映射
	switch e.Code {
	case 20000:
		return http.StatusOK
	case 40003: // ErrUnauthorized
		return http.StatusUnauthorized
	case 40004: // ErrForbidden
		return http.StatusForbidden
	case 40005: // ErrNotFound
		return http.StatusNotFound
	case 40006: // ErrTooManyReq
		return http.StatusTooManyRequests
	// Token 相关错误返回 401
	case 40104, 40105, 40106: // ErrTokenInvalid, ErrTokenExpired, ErrTokenBlacklisted
		return http.StatusUnauthorized
	}

	// 按区间映射通用错误
	switch {
	case e.Code >= 40000 && e.Code < 50000:
		return http.StatusBadRequest
	case e.Code >= 50000:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}

// Is 判断错误是否为指定错误，实现 errors.Is 接口
// 用于错误类型的比较，基于错误码进行匹配
// 参数：
//   - target: 目标错误，用于比较
//
// 返回：如果错误码相同则返回true，否则返回false
func (e *Error) Is(target error) bool {
	if t, ok := target.(*Error); ok {
		return e.Code == t.Code
	}
	return false
}
