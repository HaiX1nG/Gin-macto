package dto

// CreateChannelRequest 创建频道请求
type CreateChannelRequest struct {
	Name      string  `json:"name" binding:"required,min=1,max=100"`
	Type      int8    `json:"type" binding:"required,min=1,max=3"`
	Topic     string  `json:"topic" binding:"omitempty,max=500"`
	ParentID  *uint64 `json:"parentId"`
	Bitrate   int     `json:"bitrate" binding:"omitempty"`
	UserLimit int     `json:"userLimit" binding:"omitempty"`
	SlowMode  int     `json:"slowMode" binding:"omitempty"`
}

// UpdateChannelRequest 更新频道请求
type UpdateChannelRequest struct {
	Name      string  `json:"name" binding:"omitempty,min=1,max=100"`
	Topic     string  `json:"topic" binding:"omitempty,max=500"`
	ParentID  *uint64 `json:"parentId"`
	Position  *int    `json:"position"`
	Bitrate   *int    `json:"bitrate"`
	UserLimit *int    `json:"userLimit"`
	SlowMode  *int    `json:"slowMode"`
}

// ReorderChannelsRequest 频道排序请求
type ReorderChannelsRequest struct {
	Orders []struct {
		ID       uint64 `json:"id"`
		Position int    `json:"position"`
	} `json:"orders" binding:"required,min=1"`
}
