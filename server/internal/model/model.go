package model

import "time"

// User 用户表模型
// 存储用户基本信息，包括用户名、密码哈希、邮箱和头像等
type User struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                      // ID 用户唯一标识，自增主键
	Username     string    `gorm:"column:username;type:varchar(50);not null;uniqueIndex:uk_username" json:"username"` // Username 用户名，唯一，最大50字符
	PasswordHash string    `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`                          // PasswordHash 密码哈希值，不返回给前端
	Email        string    `gorm:"column:email;type:varchar(100);uniqueIndex:uk_email" json:"email"`                  // Email 用户邮箱，唯一，最大100字符
	AvatarURL    string    `gorm:"column:avatar_url;type:varchar(500)" json:"avatarUrl"`                              // AvatarURL 用户头像URL，最大500字符
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`                                 // CreatedAt 创建时间，自动填充
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`                                 // UpdatedAt 更新时间，自动更新
}

// TableName 返回表名
func (User) TableName() string {
	return "users"
}

// Room 房间表模型
// 存储房间基本信息，包括房间名称、类型、房主、隐私设置等
type Room struct {
	ID                    uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                     // ID 房间唯一标识，自增主键
	RoomName              string    `gorm:"column:room_name;type:varchar(100);not null" json:"roomName"`                      // RoomName 房间名称，最大100字符
	RoomType              int8      `gorm:"column:room_type;type:tinyint;not null" json:"roomType"`                           // RoomType 房间类型：1=文字聊天房，2=语音房
	HostUserID            uint64    `gorm:"column:host_user_id;not null;index:idx_host" json:"hostUserId"`                    // HostUserID 房主用户ID
	IsPrivate             bool      `gorm:"column:is_private;type:tinyint(1);not null;default:0" json:"isPrivate"`            // IsPrivate 是否为私密房间
	InviteCode            *string   `gorm:"column:invite_code;type:varchar(20);uniqueIndex:uk_invite_code" json:"inviteCode"` // InviteCode 邀请码，私密房间使用，唯一
	MaxParticipants       uint32    `gorm:"column:max_participants;not null;default:20" json:"maxParticipants"`               // MaxParticipants 最大参与者数量，默认20
	CurrentPlaylistItemID *uint64   `gorm:"column:current_playlist_item_id" json:"currentPlaylistItemId"`                     // CurrentPlaylistItemID 当前播放项ID
	CreatedAt             time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`                                // CreatedAt 创建时间
	UpdatedAt             time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`                                // UpdatedAt 更新时间
}

// TableName 返回表名
func (Room) TableName() string {
	return "rooms"
}

// RoomParticipant 房间参与者表模型
// 记录用户在房间中的参与状态、角色和实时状态信息
type RoomParticipant struct {
	ID              uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                                        // ID 参与记录唯一标识，自增主键
	RoomID          uint64     `gorm:"column:room_id;not null;uniqueIndex:uk_room_user_active;index:idx_room" json:"roomId"`                // RoomID 房间ID
	UserID          uint64     `gorm:"column:user_id;not null;uniqueIndex:uk_room_user_active;index:idx_user" json:"userId"`                // UserID 用户ID
	Role            int8       `gorm:"column:role;type:tinyint;not null;default:2" json:"role"`                                             // Role 角色：1=房主，2=发言人，3=听众
	IsMuted         bool       `gorm:"column:is_muted;type:tinyint(1);not null;default:0" json:"isMuted"`                                   // IsMuted 是否静音
	IsScreenSharing bool       `gorm:"column:is_screen_sharing;type:tinyint(1);not null;default:0" json:"isScreenSharing"`                  // IsScreenSharing 是否正在屏幕共享
	IsActive        bool       `gorm:"column:is_active;type:tinyint(1);not null;default:1;uniqueIndex:uk_room_user_active" json:"isActive"` // IsActive 是否活跃，用于软删除
	JoinedAt        time.Time  `gorm:"column:joined_at;autoCreateTime" json:"joinedAt"`                                                     // JoinedAt 加入时间
	LeftAt          *time.Time `gorm:"column:left_at" json:"leftAt"`                                                                        // LeftAt 离开时间，为空表示仍在房间
}

// TableName 返回表名
func (RoomParticipant) TableName() string {
	return "room_participants"
}

// PlaylistItem 音乐播放列表模型
// 记录房间中的音乐播放项，包括音乐信息、播放顺序和状态
type PlaylistItem struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                      // ID 播放项唯一标识，自增主键
	RoomID    uint64    `gorm:"column:room_id;not null;index:idx_room_order;index:idx_room_status" json:"roomId"`  // RoomID 所属房间ID
	AddedBy   uint64    `gorm:"column:added_by;not null" json:"addedBy"`                                           // AddedBy 添加该音乐的用户ID
	Title     string    `gorm:"column:title;type:varchar(200);not null" json:"title"`                              // Title 音乐标题，最大200字符
	Artist    string    `gorm:"column:artist;type:varchar(200)" json:"artist"`                                     // Artist 艺术家名称，最大200字符
	MusicURL  string    `gorm:"column:music_url;type:varchar(500);not null" json:"musicUrl"`                       // MusicURL 音乐文件URL，最大500字符
	Duration  uint32    `gorm:"column:duration" json:"duration"`                                                   // Duration 音乐时长（秒）
	PlayOrder uint32    `gorm:"column:play_order;not null;default:0;index:idx_room_order" json:"playOrder"`        // PlayOrder 播放顺序
	Status    int8      `gorm:"column:status;type:tinyint;not null;default:0;index:idx_room_status" json:"status"` // Status 播放状态：0=等待，1=播放中，2=已播放
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`                                 // CreatedAt 创建时间
}

// TableName 返回表名
func (PlaylistItem) TableName() string {
	return "playlist_items"
}

// ChatMessage 聊天消息表模型
// 记录房间内的聊天消息，支持文本、表情和系统消息
type ChatMessage struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                           // ID 消息唯一标识，自增主键
	RoomID       uint64    `gorm:"column:room_id;not null;index:idx_room_time" json:"roomId"`              // RoomID 所属房间ID
	SenderUserID uint64    `gorm:"column:sender_user_id;not null" json:"senderUserId"`                     // SenderUserID 发送者用户ID
	MessageType  int8      `gorm:"column:message_type;type:tinyint;not null;default:1" json:"messageType"` // MessageType 消息类型：1=文本，2=表情，3=系统消息
	Content      string    `gorm:"column:content;type:text;not null" json:"content"`                       // Content 消息内容
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime;index:idx_room_time" json:"createdAt"`  // CreatedAt 发送时间
}

// TableName 返回表名
func (ChatMessage) TableName() string {
	return "chat_messages"
}

// ScreenShareSession 屏幕共享会话表模型
// 记录房间内的屏幕共享会话，包括开始和结束时间
type ScreenShareSession struct {
	ID        uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`              // ID 会话唯一标识，自增主键
	RoomID    uint64     `gorm:"column:room_id;not null;index:idx_room_user" json:"roomId"` // RoomID 所属房间ID
	UserID    uint64     `gorm:"column:user_id;not null;index:idx_room_user" json:"userId"` // UserID 共享屏幕的用户ID
	StartedAt time.Time  `gorm:"column:started_at;not null" json:"startedAt"`               // StartedAt 开始共享时间
	EndedAt   *time.Time `gorm:"column:ended_at" json:"endedAt"`                            // EndedAt 结束共享时间，为空表示正在共享
}

// TableName 返回表名
func (ScreenShareSession) TableName() string {
	return "screen_share_sessions"
}

// VoiceSession 语音会话表模型
// 记录用户在房间内的语音参与状态，包括加入和离开时间
type VoiceSession struct {
	ID       uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`              // ID 会话唯一标识，自增主键
	RoomID   uint64     `gorm:"column:room_id;not null;index:idx_room_user" json:"roomId"` // RoomID 所属房间ID
	UserID   uint64     `gorm:"column:user_id;not null;index:idx_room_user" json:"userId"` // UserID 参与语音的用户ID
	JoinedAt time.Time  `gorm:"column:joined_at;not null" json:"joinedAt"`                 // JoinedAt 加入语音时间
	LeftAt   *time.Time `gorm:"column:left_at" json:"leftAt"`                              // LeftAt 离开语音时间，为空表示仍在语音中
}

// TableName 返回表名
func (VoiceSession) TableName() string {
	return "voice_sessions"
}

// UserStatus 用户在线状态表模型
// 记录用户的在线状态、自定义状态和最后活跃时间
type UserStatus struct {
	ID           uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                          // ID 状态记录唯一标识，自增主键
	UserID       uint64     `gorm:"column:user_id;not null;uniqueIndex:uk_user" json:"userId"`             // UserID 用户ID，唯一
	IsOnline     bool       `gorm:"column:is_online;type:tinyint(1);not null;default:0" json:"isOnline"`   // IsOnline 是否在线
	CustomStatus string     `gorm:"column:custom_status;type:varchar(100);default:''" json:"customStatus"` // CustomStatus 用户自定义状态，最大100字符
	LastSeenAt   *time.Time `gorm:"column:last_seen_at" json:"lastSeenAt"`                                 // LastSeenAt 最后活跃时间
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`                     // CreatedAt 创建时间
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`                     // UpdatedAt 更新时间
}

// TableName 返回表名
func (UserStatus) TableName() string {
	return "user_status"
}
