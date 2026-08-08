package dto

// JoinVoiceRequest 加入语音频道请求
type JoinVoiceRequest struct {
	// 无需额外参数，channelId 从路径获取
}

// VoiceParticipantResponse 语音参与者响应
type VoiceParticipantResponse struct {
	ID         uint64 `json:"id"`
	ChannelID  uint64 `json:"channelId"`
	UserID     uint64 `json:"userId"`
	Username   string `json:"username"`
	AvatarURL  string `json:"avatarUrl"`
	IsMuted    bool   `json:"isMuted"`
	IsDeafened bool   `json:"isDeafened"`
	IsSpeaking bool   `json:"isSpeaking"`
	Volume     int    `json:"volume"`
	JoinedAt   string `json:"joinedAt"`
}

// SetMuteRequest 设置静音请求
type SetMuteRequest struct {
	IsMuted bool `json:"isMuted"`
}

// ScreenShareResponse 屏幕共享响应
type ScreenShareResponse struct {
	ID        uint64 `json:"id"`
	ChannelID uint64 `json:"channelId"`
	UserID    uint64 `json:"userId"`
	Username  string `json:"username"`
	StartedAt string `json:"startedAt"`
}

// WebRTCSignalRequest WebRTC信令请求（保留，REST 备用通道）
type WebRTCSignalRequest struct {
	Type     string `json:"type"`     // offer, answer, ice-candidate
	TargetID uint64 `json:"targetId"` // 目标用户ID
	Payload  string `json:"payload"`  // SDP或ICE候选数据
}
