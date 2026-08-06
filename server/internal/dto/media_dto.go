package dto

// AddPlaylistItemRequest 添加播放项请求
type AddPlaylistItemRequest struct {
	Title    string `json:"title" binding:"required,min=1,max=200"`
	Artist   string `json:"artist" binding:"omitempty,max=200"`
	MusicURL string `json:"musicUrl" binding:"required,url,max=500"`
	Duration uint32 `json:"duration" binding:"omitempty"`
}

// PlaylistItemResponse 播放项响应
type PlaylistItemResponse struct {
	ID        uint64 `json:"id"`
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	MusicURL  string `json:"musicUrl"`
	Duration  uint32 `json:"duration"`
	PlayOrder uint32 `json:"playOrder"`
	Status    int8   `json:"status"`
	AddedBy   uint64 `json:"addedBy"`
}

// ReorderPlaylistRequest 调整播放顺序请求
type ReorderPlaylistRequest struct {
	ItemIDs []uint64 `json:"itemIds" binding:"required,min=1"`
}

// UpdateMessageRequest 编辑消息请求
// 对应前端 chatService.updateMessage(roomId, messageId, content)
type UpdateMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	MessageType int8   `json:"messageType" binding:"required,min=1,max=3"`
	Content     string `json:"content" binding:"required,min=1,max=2000"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	ID           uint64 `json:"id"`
	RoomID       uint64 `json:"roomId"`
	SenderUserID uint64 `json:"senderUserId"`
	SenderName   string `json:"senderName"`
	MessageType  int8   `json:"messageType"`
	Content      string `json:"content"`
	CreatedAt    string `json:"createdAt"`
}

// MessageListRequest 消息列表请求
type MessageListRequest struct {
	Page        int    `form:"page" binding:"omitempty,min=1"`
	PageSize    int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	MessageType int8   `form:"messageType" binding:"omitempty,min=1,max=3"`
	SenderID    uint64 `form:"senderId"`
	StartTime   string `form:"startTime"` // 格式: 2006-01-02 15:04:05
	EndTime     string `form:"endTime"`   // 格式: 2006-01-02 15:04:05
}

// UserHistoryMessagesRequest 用户历史消息请求
type UserHistoryMessagesRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

// UserHistoryMessageResponse 用户历史消息响应
type UserHistoryMessageResponse struct {
	ID           uint64 `json:"id"`
	RoomID       uint64 `json:"roomId"`
	RoomName     string `json:"roomName"`
	SenderUserID uint64 `json:"senderUserId"`
	SenderName   string `json:"senderName"`
	MessageType  int8   `json:"messageType"`
	Content      string `json:"content"`
	CreatedAt    string `json:"createdAt"`
}

// SearchMessagesRequest 消息搜索请求
type SearchMessagesRequest struct {
	Query    string `form:"query" binding:"required,min=1,max=100"`
	RoomID   uint64 `form:"roomId" binding:"omitempty"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

// SearchMessagesResponse 消息搜索响应
type SearchMessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
	Total    int64             `json:"total"`
}
