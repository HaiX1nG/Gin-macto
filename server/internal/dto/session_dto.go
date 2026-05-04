package dto

// StartScreenShareRequest 开始屏幕共享请求
type StartScreenShareRequest struct {
	// 无需额外参数，从token获取用户ID
}

// ScreenShareResponse 屏幕共享响应
type ScreenShareResponse struct {
	ID        uint64 `json:"id"`
	RoomID    uint64 `json:"roomId"`
	UserID    uint64 `json:"userId"`
	Username  string `json:"username"`
	StartedAt string `json:"startedAt"`
}

// StartVoiceRequest 开始语音请求
type StartVoiceRequest struct {
	// 无需额外参数
}

// VoiceSessionResponse 语音会话响应
type VoiceSessionResponse struct {
	ID       uint64 `json:"id"`
	RoomID   uint64 `json:"roomId"`
	UserID   uint64 `json:"userId"`
	Username string `json:"username"`
	JoinedAt string `json:"joinedAt"`
}

// WebRTCSignalRequest WebRTC信令请求
type WebRTCSignalRequest struct {
	Type     string `json:"type"`     // offer, answer, ice-candidate
	TargetID uint64 `json:"targetId"` // 目标用户ID（可选，用于点对点）
	Payload  string `json:"payload"`  // SDP或ICE候选数据
}
