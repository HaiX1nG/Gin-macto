package model

import "time"

// 消息类型枚举
const (
	MessageTypeText   int8 = 1 // 文本消息
	MessageTypeImage  int8 = 2 // 图片消息
	MessageTypeSystem int8 = 3 // 系统消息
)

// ChannelMessage 频道消息表模型
// 记录频道内的聊天消息，支持文本、图片和系统消息
type ChannelMessage struct {
	ID           uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                               // ID 消息唯一标识，自增主键
	ChannelID    uint64     `gorm:"column:channel_id;not null;index:idx_channel_created" json:"channelId"`       // ChannelID 所属频道ID
	SenderUserID uint64     `gorm:"column:sender_user_id;not null" json:"senderUserId"`                         // SenderUserID 发送者用户ID
	Type         int8       `gorm:"column:type;type:tinyint;not null;default:1" json:"type"`                     // Type 消息类型：1=文本，2=图片，3=系统
	Content      string     `gorm:"column:content;type:text;not null" json:"content"`                            // Content 消息内容
	ReplyToID    *uint64    `gorm:"column:reply_to_id" json:"replyToId"`                                         // ReplyToID 回复的消息ID，null=非回复
	EditedAt     *time.Time `gorm:"column:edited_at" json:"editedAt"`                                            // EditedAt 编辑时间，null=未编辑
	IsPinned     bool       `gorm:"column:is_pinned;type:tinyint(1);not null;default:0" json:"isPinned"`        // IsPinned 是否置顶
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime;index:idx_channel_created" json:"createdAt"` // CreatedAt 发送时间
}

// TableName 返回表名
func (ChannelMessage) TableName() string {
	return "channel_messages"
}

// MessageAttachment 消息附件表模型
// 记录消息的附件信息，如图片、文件等
type MessageAttachment struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                    // ID 附件唯一标识，自增主键
	MessageID uint64    `gorm:"column:message_id;not null;index:idx_message" json:"messageId"`   // MessageID 所属消息ID
	Filename  string    `gorm:"column:filename;type:varchar(255)" json:"filename"`              // Filename 文件名，最大255字符
	URL       string    `gorm:"column:url;type:varchar(500)" json:"url"`                        // URL 文件访问URL
	FileSize  int       `gorm:"column:file_size" json:"fileSize"`                                // FileSize 文件大小（字节）
	MimeType  string    `gorm:"column:mime_type;type:varchar(100)" json:"mimeType"`             // MimeType MIME类型，最大100字符
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`               // CreatedAt 创建时间
}

// TableName 返回表名
func (MessageAttachment) TableName() string {
	return "message_attachments"
}

// MessageReaction 消息表情反应表模型
// 记录用户对消息的表情反应，每个用户对每条消息的每个表情只能反应一次
type MessageReaction struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                                      // ID 反应记录唯一标识，自增主键
	MessageID uint64    `gorm:"column:message_id;not null;uniqueIndex:uk_message_user_emoji;index:idx_message" json:"messageId"`   // MessageID 所属消息ID
	UserID    uint64    `gorm:"column:user_id;not null;uniqueIndex:uk_message_user_emoji" json:"userId"`                            // UserID 反应用户ID
	Emoji     string    `gorm:"column:emoji;type:varchar(32);not null;uniqueIndex:uk_message_user_emoji" json:"emoji"`              // Emoji emoji字符或自定义标识，最大32字符
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`                                                  // CreatedAt 创建时间
}

// TableName 返回表名
func (MessageReaction) TableName() string {
	return "message_reactions"
}
