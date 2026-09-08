// Package dto 定义数据传输对象（DTO）
// WebRTC 信令类型定义，用于前端状态同步
package dto

// ConnectionState WebRTC 连接状态枚举
// 前端根据此状态显示对应的 UI 反馈
type ConnectionState string

const (
	// ConnectionStateNew 新建连接
	ConnectionStateNew ConnectionState = "new"
	// ConnectionStateConnecting 正在建立连接
	ConnectionStateConnecting ConnectionState = "connecting"
	// ConnectionStateConnected 已连接
	ConnectionStateConnected ConnectionState = "connected"
	// ConnectionStateDisconnected 已断开
	ConnectionStateDisconnected ConnectionState = "disconnected"
	// ConnectionStateFailed 连接失败
	ConnectionStateFailed ConnectionState = "failed"
	// ConnectionStateClosed 已关闭
	ConnectionStateClosed ConnectionState = "closed"
)

// ConnectionFailureCode 连接失败原因枚举
// 前端根据此代码显示对应的用户提示
type ConnectionFailureCode string

const (
	// FailureCodeNone 无失败
	FailureCodeNone ConnectionFailureCode = ""
	// FailureCodeSignalTimeout 信令超时
	FailureCodeSignalTimeout ConnectionFailureCode = "signal_timeout"
	// FailureCodePeerBusy 对方忙
	FailureCodePeerBusy ConnectionFailureCode = "peer_busy"
	// FailureCodePermissionDenied 权限拒绝（麦克风/屏幕录制）
	FailureCodePermissionDenied ConnectionFailureCode = "permission_denied"
	// FailureCodeNetworkError 网络错误
	FailureCodeNetworkError ConnectionFailureCode = "network_error"
	// FailureCodeServerError 服务器错误
	FailureCodeServerError ConnectionFailureCode = "server_error"
)

// WebRTCSignalResponse WebRTC 信令响应
type WebRTCSignalResponse struct {
	// Success 是否成功
	Success bool `json:"success"`
	// Message 提示信息
	Message string `json:"message"`
}

// ConnectionStatePush 连接状态推送
// 服务器在连接状态变化时主动推送给前端
type ConnectionStatePush struct {
	// Status 当前连接状态
	Status ConnectionState `json:"status"`
	// ReconnectAttempt 当前重连次数
	ReconnectAttempt int `json:"reconnectAttempt,omitempty"`
	// MaxReconnectAttempts 最大重连次数
	MaxReconnectAttempts int `json:"maxReconnectAttempts,omitempty"`
	// FailureCode 失败原因码
	FailureCode ConnectionFailureCode `json:"failureCode,omitempty"`
	// FailureMessage 失败提示信息
	FailureMessage string `json:"failureMessage,omitempty"`
}
