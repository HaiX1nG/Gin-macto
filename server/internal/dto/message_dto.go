package dto

import "time"

// SendMessageRequest 发送频道消息请求
type SendMessageRequest struct {
	Type       int8    `json:"type" binding:"required,min=1,max=3"`
	Content    string  `json:"content" binding:"required,min=1,max=2000"`
	ReplyToID  *uint64 `json:"replyToId"`
}

// UpdateMessageRequest 编辑消息请求
type UpdateMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

// MessageListRequest 消息列表请求
type MessageListRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

// ReactionRequest 表情反应请求
type ReactionRequest struct {
	Emoji string `json:"emoji" binding:"required,min=1,max=32"`
}

// ReactionResponse 表情反应响应
type ReactionResponse struct {
	Emoji  string   `json:"emoji"`
	Count  int      `json:"count"`
	UserIDs []uint64 `json:"userIds"`
}

// AttachmentResponse 附件响应
type AttachmentResponse struct {
	ID        uint64 `json:"id"`
	Filename  string `json:"filename"`
	URL       string `json:"url"`
	FileSize  int    `json:"fileSize"`
	MimeType  string `json:"mimeType"`
}

// MessageResponse 频道消息响应
type MessageResponse struct {
	ID           uint64               `json:"id"`
	ChannelID    uint64               `json:"channelId"`
	SenderUserID uint64               `json:"senderUserId"`
	SenderName   string               `json:"senderName"`
	SenderAvatar string               `json:"senderAvatar"`
	Type         int8                 `json:"type"`
	Content      string               `json:"content"`
	ReplyToID    *uint64              `json:"replyToId,omitempty"`
	EditedAt     *time.Time           `json:"editedAt,omitempty"`
	IsPinned     bool                 `json:"isPinned"`
	Reactions    []ReactionResponse   `json:"reactions"`
	Attachments  []AttachmentResponse `json:"attachments"`
	CreatedAt    string               `json:"createdAt"`
}

// SearchMessagesRequest 消息搜索请求
type SearchMessagesRequest struct {
	Query     string  `form:"query" binding:"required,min=1,max=100"`
	ChannelID *uint64 `form:"channelId"`
	Page      int     `form:"page" binding:"omitempty,min=1"`
	PageSize  int     `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

// SearchMessagesResponse 消息搜索响应
type SearchMessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
	Total    int64             `json:"total"`
}
