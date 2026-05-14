package model

import "time"

// FriendRequestStatus 好友请求状态枚举类型
// 用于表示好友请求的处理状态
type FriendRequestStatus int

const (
	// FriendRequestPending 待处理状态，好友请求尚未被接受或拒绝
	FriendRequestPending FriendRequestStatus = 0
	// FriendRequestAccepted 已接受状态，好友请求已被接收方接受
	FriendRequestAccepted FriendRequestStatus = 1
	// FriendRequestRejected 已拒绝状态，好友请求已被接收方拒绝
	FriendRequestRejected FriendRequestStatus = 2
)

// FriendRequest 好友请求表模型
// 记录好友请求信息，包括发送方、接收方、状态和附言
type FriendRequest struct {
	ID         uint64              `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                     // ID 好友请求唯一标识，自增主键
	SenderID   uint64              `gorm:"column:sender_id;not null;index:idx_sender" json:"senderId"`       // SenderID 发送请求的用户ID
	ReceiverID uint64              `gorm:"column:receiver_id;not null;index:idx_receiver" json:"receiverId"` // ReceiverID 接收请求的用户ID
	Status     FriendRequestStatus `gorm:"column:status;type:tinyint;not null;default:0" json:"status"`      // Status 请求状态：0=待处理，1=已接受，2=已拒绝
	Message    string              `gorm:"column:message;type:varchar(200)" json:"message"`                  // Message 请求附言，最大200字符
	CreatedAt  time.Time           `gorm:"column:created_at;autoCreateTime" json:"createdAt"`                // CreatedAt 创建时间
	UpdatedAt  time.Time           `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`                // UpdatedAt 更新时间
}

// TableName 返回表名
func (FriendRequest) TableName() string {
	return "friend_requests"
}

// Friendship 好友关系表模型
// 记录已建立的好友关系，双方各有一条记录
type Friendship struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                         // ID 好友关系唯一标识，自增主键
	UserID    uint64    `gorm:"column:user_id;not null;uniqueIndex:uk_user_friend" json:"userId"`     // UserID 用户ID，与FriendID组成唯一索引
	FriendID  uint64    `gorm:"column:friend_id;not null;uniqueIndex:uk_user_friend" json:"friendId"` // FriendID 好友用户ID，与UserID组成唯一索引
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`                    // CreatedAt 好友关系建立时间
}

// TableName 返回表名
func (Friendship) TableName() string {
	return "friendships"
}

// PrivateMessage 私聊消息表模型
// 记录用户之间的私聊消息，支持消息状态跟踪
type PrivateMessage struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                             // ID 消息唯一标识，自增主键
	SenderID   uint64    `gorm:"column:sender_id;not null;index:idx_sender" json:"senderId"`               // SenderID 发送者用户ID
	ReceiverID uint64    `gorm:"column:receiver_id;not null;index:idx_receiver" json:"receiverId"`         // ReceiverID 接收者用户ID
	Content    string    `gorm:"column:content;type:text;not null" json:"content"`                         // Content 消息内容
	IsRead     bool      `gorm:"column:is_read;type:tinyint(1);not null;default:0" json:"isRead"`          // IsRead 是否已读
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime;index:idx_conversation" json:"createdAt"` // CreatedAt 发送时间
}

// TableName 返回表名
func (PrivateMessage) TableName() string {
	return "private_messages"
}
