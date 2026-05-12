package dto

// SendFriendRequestRequest 发送好友请求
type SendFriendRequestRequest struct {
	ReceiverID uint64 `json:"receiverId" binding:"required"`
	Message    string `json:"message" binding:"omitempty,max=200"`
}

// HandleFriendRequestRequest 处理好友请求
type HandleFriendRequestRequest struct {
	Accept bool `json:"accept" binding:"required"`
}

// FriendRequestResponse 好友请求响应
type FriendRequestResponse struct {
	ID         uint64 `json:"id"`
	SenderID   uint64 `json:"senderId"`
	SenderName string `json:"senderName"`
	ReceiverID uint64 `json:"receiverId"`
	Status     int    `json:"status"` // 0=pending, 1=accepted, 2=rejected
	Message    string `json:"message"`
	CreatedAt  string `json:"createdAt"`
}

// FriendResponse 好友响应
type FriendResponse struct {
	FriendID     uint64 `json:"friendId"`
	FriendName   string `json:"friendName"`
	AvatarURL    string `json:"avatarUrl"`
	IsOnline     bool   `json:"isOnline"`
	CustomStatus string `json:"customStatus"`
	CreatedAt    string `json:"createdAt"`
}

// FriendListResponse 好友列表响应
type FriendListResponse struct {
	Friends []FriendResponse `json:"friends"`
	Total   int64            `json:"total"`
}

// PendingRequestsResponse 待处理请求响应
type PendingRequestsResponse struct {
	Requests []FriendRequestResponse `json:"requests"`
	Total    int64                   `json:"total"`
}

// SendPrivateMessageRequest 发送私聊消息请求
type SendPrivateMessageRequest struct {
	ReceiverID uint64 `json:"receiverId" binding:"required"`
	Content    string `json:"content" binding:"required,min=1,max=5000"`
}

// PrivateMessageResponse 私聊消息响应
type PrivateMessageResponse struct {
	ID         uint64 `json:"id"`
	SenderID   uint64 `json:"senderId"`
	SenderName string `json:"senderName"`
	ReceiverID uint64 `json:"receiverId"`
	Content    string `json:"content"`
	IsRead     bool   `json:"isRead"`
	CreatedAt  string `json:"createdAt"`
}

// PrivateMessageListResponse 私聊消息列表响应
type PrivateMessageListResponse struct {
	Messages []PrivateMessageResponse `json:"messages"`
	Total    int64                    `json:"total"`
}

// ConversationResponse 会话响应
type ConversationResponse struct {
	UserID       uint64 `json:"userId"`
	Username     string `json:"username"`
	AvatarURL    string `json:"avatarUrl"`
	IsOnline     bool   `json:"isOnline"`
	CustomStatus string `json:"customStatus"`
	LastMessage  string `json:"lastMessage"`
	UnreadCount  int64  `json:"unreadCount"`
}

// ConversationListResponse 会话列表响应
type ConversationListResponse struct {
	Conversations []ConversationResponse `json:"conversations"`
	Total         int64                  `json:"total"`
}

// SearchUserResponse 搜索用户响应
type SearchUserResponse struct {
	UserID            uint64 `json:"userId"`
	Username          string `json:"username"`
	AvatarURL         string `json:"avatarUrl"`
	IsOnline          bool   `json:"isOnline"`
	CustomStatus      string `json:"customStatus"`
	IsFriend          bool   `json:"isFriend"`
	HasPendingRequest bool   `json:"hasPendingRequest"`
}
