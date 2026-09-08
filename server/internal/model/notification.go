package model

import "time"

// NotificationType 通知类型枚举
// 用于区分不同类型的通知消息
type NotificationType string

const (
	// NotificationTypeMention 提及通知类型，当用户在频道中被@提及时触发
	NotificationTypeMention NotificationType = "mention"
	// NotificationTypePin 置顶通知类型，当频道中有消息被置顶时触发
	NotificationTypePin NotificationType = "pin"
	// NotificationTypeFriendRequest 好友请求通知类型，当收到好友请求时触发
	NotificationTypeFriendRequest NotificationType = "friend_request"
)

// Notification 通知表模型
// 记录用户收到的各类通知消息，支持多种通知类型
type Notification struct {
	ID        uint64           `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                        // ID 通知唯一标识，自增主键
	UserID    uint64           `gorm:"column:user_id;not null;index:idx_user" json:"userId"`                                // UserID 接收通知的用户ID
	Type      NotificationType `gorm:"column:type;type:varchar(30);not null;index:idx_user_type" json:"type"`               // Type 通知类型：mention/pin/friend_request
	SourceID  uint64           `gorm:"column:source_id;not null" json:"sourceId"`                                           // SourceID 来源ID（消息ID/请求ID等）
	Title     string           `gorm:"column:title;type:varchar(200);not null" json:"title"`                                // Title 通知标题，最大200字符
	Content   string           `gorm:"column:content;type:text" json:"content"`                                             // Content 通知正文内容
	IsRead    bool             `gorm:"column:is_read;type:tinyint(1);not null;default:0;index:idx_user_read" json:"isRead"` // IsRead 是否已读
	CreatedAt time.Time        `gorm:"column:created_at;autoCreateTime;index:idx_user" json:"createdAt"`                    // CreatedAt 通知创建时间
}

// TableName 返回表名
func (Notification) TableName() string {
	return "notifications"
}
