package model

import "time"

// VoiceParticipant 语音参与者实时状态表模型
// 记录当前在语音频道中的用户及其实时状态（静音、说话等）
type VoiceParticipant struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                              // ID 参与者记录唯一标识，自增主键
	ChannelID   uint64    `gorm:"column:channel_id;not null;uniqueIndex:uk_channel_user;index:idx_channel" json:"channelId"` // ChannelID 所属语音频道ID
	UserID      uint64    `gorm:"column:user_id;not null;uniqueIndex:uk_channel_user" json:"userId"`                          // UserID 参与语音的用户ID
	IsMuted     bool      `gorm:"column:is_muted;type:tinyint(1);not null;default:0" json:"isMuted"`                          // IsMuted 是否静音
	IsDeafened  bool      `gorm:"column:is_deafened;type:tinyint(1);not null;default:0" json:"isDeafened"`                    // IsDeafened 是否消音
	IsSpeaking  bool      `gorm:"column:is_speaking;type:tinyint(1);not null;default:0" json:"isSpeaking"`                     // IsSpeaking 是否正在说话
	Volume      int       `gorm:"column:volume;not null;default:100" json:"volume"`                                           // Volume 音量，默认100
	JoinedAt    time.Time `gorm:"column:joined_at;autoCreateTime" json:"joinedAt"`                                             // JoinedAt 加入语音时间
}

// TableName 返回表名
func (VoiceParticipant) TableName() string {
	return "voice_participants"
}

// VoiceSession 语音会话历史记录表模型
// 记录用户在语音频道中的参与历史，用于统计和审计
type VoiceSession struct {
	ID        uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                              // ID 会话记录唯一标识，自增主键
	ChannelID uint64     `gorm:"column:channel_id;not null;index:idx_channel_user" json:"channelId"`        // ChannelID 所属语音频道ID
	UserID    uint64     `gorm:"column:user_id;not null;index:idx_channel_user" json:"userId"`             // UserID 参与语音的用户ID
	JoinedAt  time.Time  `gorm:"column:joined_at;not null" json:"joinedAt"`                                 // JoinedAt 加入语音时间
	LeftAt    *time.Time `gorm:"column:left_at" json:"leftAt"`                                              // LeftAt 离开语音时间，null=仍在语音中
}

// TableName 返回表名
func (VoiceSession) TableName() string {
	return "voice_sessions"
}

// ScreenShareSession 屏幕共享会话表模型
// 记录语音频道内的屏幕共享会话，包括开始和结束时间
type ScreenShareSession struct {
	ID        uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`              // ID 会话唯一标识，自增主键
	ChannelID uint64     `gorm:"column:channel_id;not null;index:idx_channel" json:"channelId"` // ChannelID 所属语音频道ID
	UserID    uint64     `gorm:"column:user_id;not null" json:"userId"`                     // UserID 共享屏幕的用户ID
	StartedAt time.Time  `gorm:"column:started_at;not null" json:"startedAt"`               // StartedAt 开始共享时间
	EndedAt   *time.Time `gorm:"column:ended_at" json:"endedAt"`                            // EndedAt 结束共享时间，null=正在共享
}

// TableName 返回表名
func (ScreenShareSession) TableName() string {
	return "screen_share_sessions"
}
