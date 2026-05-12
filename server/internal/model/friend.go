package model

import "time"

// FriendRequestStatus 好友请求状态
type FriendRequestStatus int

const (
	FriendRequestPending  FriendRequestStatus = 0 // 待处理
	FriendRequestAccepted FriendRequestStatus = 1 // 已接受
	FriendRequestRejected FriendRequestStatus = 2 // 已拒绝
)

// FriendRequest 好友请求表模型
type FriendRequest struct {
	ID         uint64              `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SenderID   uint64              `gorm:"column:sender_id;not null;index:idx_sender" json:"senderId"`
	ReceiverID uint64              `gorm:"column:receiver_id;not null;index:idx_receiver" json:"receiverId"`
	Status     FriendRequestStatus `gorm:"column:status;type:tinyint;not null;default:0" json:"status"` // 0=pending, 1=accepted, 2=rejected
	Message    string              `gorm:"column:message;type:varchar(200)" json:"message"`             // 请求附言
	CreatedAt  time.Time           `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time           `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

// TableName 返回表名
func (FriendRequest) TableName() string {
	return "friend_requests"
}

// Friendship 好友关系表模型
type Friendship struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"column:user_id;not null;uniqueIndex:uk_user_friend" json:"userId"`
	FriendID  uint64    `gorm:"column:friend_id;not null;uniqueIndex:uk_user_friend" json:"friendId"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

// TableName 返回表名
func (Friendship) TableName() string {
	return "friendships"
}

// PrivateMessage 私聊消息表模型
type PrivateMessage struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SenderID   uint64    `gorm:"column:sender_id;not null;index:idx_sender" json:"senderId"`
	ReceiverID uint64    `gorm:"column:receiver_id;not null;index:idx_receiver" json:"receiverId"`
	Content    string    `gorm:"column:content;type:text;not null" json:"content"`
	IsRead     bool      `gorm:"column:is_read;type:tinyint(1);not null;default:0" json:"isRead"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime;index:idx_conversation" json:"createdAt"`
}

// TableName 返回表名
func (PrivateMessage) TableName() string {
	return "private_messages"
}
