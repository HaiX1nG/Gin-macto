package dto

// CreateRoomRequest 创建房间请求
type CreateRoomRequest struct {
	RoomName        string `json:"roomName" binding:"required,min=1,max=100"`
	RoomType        int8   `json:"roomType" binding:"required,min=1,max=4"`
	IsPrivate       bool   `json:"isPrivate"`
	MaxParticipants uint32 `json:"maxParticipants" binding:"omitempty,min=2,max=100"`
}

// JoinRoomRequest 加入房间请求
type JoinRoomRequest struct {
	InviteCode string `json:"inviteCode"` // 私密房间需要邀请码
}

// RoomInfoResponse 房间信息响应
type RoomInfoResponse struct {
	ID                    uint64  `json:"id"`
	RoomName              string  `json:"roomName"`
	RoomType              int8    `json:"roomType"`
	HostUserID            uint64  `json:"hostUserId"`
	IsPrivate             bool    `json:"isPrivate"`
	InviteCode            string  `json:"inviteCode,omitempty"`
	MaxParticipants       uint32  `json:"maxParticipants"`
	CurrentPlaylistItemID *uint64 `json:"currentPlaylistItemId,omitempty"`
	ParticipantCount      int     `json:"participantCount"`
	CreatedAt             string  `json:"createdAt"`
}

// ParticipantResponse 参与者响应
type ParticipantResponse struct {
	UserID          uint64 `json:"userId"`
	Username        string `json:"username"`
	AvatarURL       string `json:"avatarUrl"`
	Role            int8   `json:"role"`
	IsMuted         bool   `json:"isMuted"`
	IsScreenSharing bool   `json:"isScreenSharing"`
	JoinedAt        string `json:"joinedAt"`
}

// RoomListRequest 房间列表请求
type RoomListRequest struct {
	Page     int  `form:"page" binding:"omitempty,min=1"`
	PageSize int  `form:"pageSize" binding:"omitempty,min=1,max=50"`
	RoomType int8 `form:"roomType" binding:"omitempty,min=1,max=4"`
}
