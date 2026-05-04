package model

import "time"

// User 用户表模型
type User struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"column:username;type:varchar(50);not null;uniqueIndex:uk_username" json:"username"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	Email        string    `gorm:"column:email;type:varchar(100);uniqueIndex:uk_email" json:"email"`
	AvatarURL    string    `gorm:"column:avatar_url;type:varchar(500)" json:"avatarUrl"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

// TableName 返回表名
func (User) TableName() string {
	return "users"
}

// Room 房间表模型
type Room struct {
	ID                    uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RoomName              string    `gorm:"column:room_name;type:varchar(100);not null" json:"roomName"`
	RoomType              int8      `gorm:"column:room_type;type:tinyint;not null" json:"roomType"` // 1=文字聊天房,2=语音房
	HostUserID            uint64    `gorm:"column:host_user_id;not null;index:idx_host" json:"hostUserId"`
	IsPrivate             bool      `gorm:"column:is_private;type:tinyint(1);not null;default:0" json:"isPrivate"`
	InviteCode            *string   `gorm:"column:invite_code;type:varchar(20);uniqueIndex:uk_invite_code" json:"inviteCode"`
	MaxParticipants       uint32    `gorm:"column:max_participants;not null;default:20" json:"maxParticipants"`
	CurrentPlaylistItemID *uint64   `gorm:"column:current_playlist_item_id" json:"currentPlaylistItemId"`
	CreatedAt             time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt             time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

// TableName 返回表名
func (Room) TableName() string {
	return "rooms"
}

// RoomParticipant 房间参与者表模型
type RoomParticipant struct {
	ID              uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RoomID          uint64     `gorm:"column:room_id;not null;uniqueIndex:uk_room_user_active;index:idx_room" json:"roomId"`
	UserID          uint64     `gorm:"column:user_id;not null;uniqueIndex:uk_room_user_active;index:idx_user" json:"userId"`
	Role            int8       `gorm:"column:role;type:tinyint;not null;default:2" json:"role"` // 1=房主,2=发言人,3=听众
	IsMuted         bool       `gorm:"column:is_muted;type:tinyint(1);not null;default:0" json:"isMuted"`
	IsScreenSharing bool       `gorm:"column:is_screen_sharing;type:tinyint(1);not null;default:0" json:"isScreenSharing"`
	IsActive        bool       `gorm:"column:is_active;type:tinyint(1);not null;default:1;uniqueIndex:uk_room_user_active" json:"isActive"`
	JoinedAt        time.Time  `gorm:"column:joined_at;autoCreateTime" json:"joinedAt"`
	LeftAt          *time.Time `gorm:"column:left_at" json:"leftAt"`
}

// TableName 返回表名
func (RoomParticipant) TableName() string {
	return "room_participants"
}

// PlaylistItem 音乐播放列表模型
type PlaylistItem struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RoomID    uint64    `gorm:"column:room_id;not null;index:idx_room_order;index:idx_room_status" json:"roomId"`
	AddedBy   uint64    `gorm:"column:added_by;not null" json:"addedBy"`
	Title     string    `gorm:"column:title;type:varchar(200);not null" json:"title"`
	Artist    string    `gorm:"column:artist;type:varchar(200)" json:"artist"`
	MusicURL  string    `gorm:"column:music_url;type:varchar(500);not null" json:"musicUrl"`
	Duration  uint32    `gorm:"column:duration" json:"duration"` // 时长（秒）
	PlayOrder uint32    `gorm:"column:play_order;not null;default:0;index:idx_room_order" json:"playOrder"`
	Status    int8      `gorm:"column:status;type:tinyint;not null;default:0;index:idx_room_status" json:"status"` // 0=等待,1=播放中,2=已播
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

// TableName 返回表名
func (PlaylistItem) TableName() string {
	return "playlist_items"
}

// ChatMessage 聊天消息表模型
type ChatMessage struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RoomID       uint64    `gorm:"column:room_id;not null;index:idx_room_time" json:"roomId"`
	SenderUserID uint64    `gorm:"column:sender_user_id;not null" json:"senderUserId"`
	MessageType  int8      `gorm:"column:message_type;type:tinyint;not null;default:1" json:"messageType"` // 1=文本,2=表情,3=系统消息
	Content      string    `gorm:"column:content;type:text;not null" json:"content"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime;index:idx_room_time" json:"createdAt"`
}

// TableName 返回表名
func (ChatMessage) TableName() string {
	return "chat_messages"
}

// ScreenShareSession 屏幕共享会话表模型
type ScreenShareSession struct {
	ID        uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RoomID    uint64     `gorm:"column:room_id;not null;index:idx_room_user" json:"roomId"`
	UserID    uint64     `gorm:"column:user_id;not null;index:idx_room_user" json:"userId"`
	StartedAt time.Time  `gorm:"column:started_at;not null" json:"startedAt"`
	EndedAt   *time.Time `gorm:"column:ended_at" json:"endedAt"`
}

// TableName 返回表名
func (ScreenShareSession) TableName() string {
	return "screen_share_sessions"
}

// VoiceSession 语音会话表模型
type VoiceSession struct {
	ID       uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RoomID   uint64     `gorm:"column:room_id;not null;index:idx_room_user" json:"roomId"`
	UserID   uint64     `gorm:"column:user_id;not null;index:idx_room_user" json:"userId"`
	JoinedAt time.Time  `gorm:"column:joined_at;not null" json:"joinedAt"`
	LeftAt   *time.Time `gorm:"column:left_at" json:"leftAt"`
}

// TableName 返回表名
func (VoiceSession) TableName() string {
	return "voice_sessions"
}

// UserStatus 用户在线状态表模型
type UserStatus struct {
	ID           uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID       uint64     `gorm:"column:user_id;not null;uniqueIndex:uk_user" json:"userId"`
	IsOnline     bool       `gorm:"column:is_online;type:tinyint(1);not null;default:0" json:"isOnline"`
	CustomStatus string     `gorm:"column:custom_status;type:varchar(100);default:''" json:"customStatus"`
	LastSeenAt   *time.Time `gorm:"column:last_seen_at" json:"lastSeenAt"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

// TableName 返回表名
func (UserStatus) TableName() string {
	return "user_status"
}
