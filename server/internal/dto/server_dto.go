package dto

// CreateServerRequest 创建服务器请求
type CreateServerRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	IconURL     string `json:"iconUrl" binding:"omitempty,url,max=500"`
	BannerURL   string `json:"bannerUrl" binding:"omitempty,url,max=500"`
	Description string `json:"description" binding:"omitempty,max=500"`
	IsPrivate   bool   `json:"isPrivate"`
}

// UpdateServerRequest 更新服务器请求
type UpdateServerRequest struct {
	Name        string `json:"name" binding:"omitempty,min=1,max=100"`
	IconURL     string `json:"iconUrl" binding:"omitempty,url,max=500"`
	BannerURL   string `json:"bannerUrl" binding:"omitempty,url,max=500"`
	Description string `json:"description" binding:"omitempty,max=500"`
}

// JoinServerRequest 加入服务器请求
type JoinServerRequest struct {
	InviteCode string `json:"inviteCode" binding:"required"`
}

// ServerInfoResponse 服务器信息响应
type ServerInfoResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	IconURL     string `json:"iconUrl"`
	BannerURL   string `json:"bannerUrl"`
	Description string `json:"description"`
	OwnerID     uint64 `json:"ownerId"`
	InviteCode  string `json:"inviteCode"`
	IsPrivate   bool   `json:"isPrivate"`
	MaxMembers  int    `json:"maxMembers"`
	MemberCount int    `json:"memberCount"`
	CreatedAt   string `json:"createdAt"`
}

// ChannelResponse 频道响应（用于服务器详情）
type ChannelResponse struct {
	ID        uint64  `json:"id"`
	ServerID  uint64  `json:"serverId"`
	Name      string  `json:"name"`
	Type      int8    `json:"type"`
	Topic     string  `json:"topic"`
	ParentID  *uint64 `json:"parentId"`
	Position  int     `json:"position"`
	Bitrate   int     `json:"bitrate"`
	UserLimit int     `json:"userLimit"`
	SlowMode  int     `json:"slowMode"`
}

// ServerDetailResponse 服务器详情响应（含频道列表）
type ServerDetailResponse struct {
	ServerInfoResponse
	Channels []ChannelResponse `json:"channels"`
}

// ServerMemberResponse 服务器成员响应
type ServerMemberResponse struct {
	ID        uint64      `json:"id"`
	ServerID  uint64      `json:"serverId"`
	UserID    uint64      `json:"userId"`
	Username  string      `json:"username"`
	AvatarURL string      `json:"avatarUrl"`
	Nickname  string      `json:"nickname"`
	JoinedAt  string      `json:"joinedAt"`
	Roles     []RoleResponse `json:"roles"`
}

// UpdateServerMemberRequest 更新服务器成员请求
type UpdateServerMemberRequest struct {
	Nickname string   `json:"nickname" binding:"omitempty,max=50"`
	RoleIDs  []uint64 `json:"roleIds"`
}
