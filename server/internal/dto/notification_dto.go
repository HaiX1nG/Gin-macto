package dto

// NotificationResponse 通知响应
// 用于返回通知数据的响应结构体
type NotificationResponse struct {
	ID        uint64 `json:"id"`        // ID 通知唯一标识
	Type      string `json:"type"`      // Type 通知类型：mention/pin/friend_request
	SourceID  uint64 `json:"sourceId"`  // SourceID 来源ID（消息ID/请求ID等）
	Title     string `json:"title"`     // Title 通知标题
	Content   string `json:"content"`   // Content 通知正文内容
	IsRead    bool   `json:"isRead"`    // IsRead 是否已读
	CreatedAt string `json:"createdAt"` // CreatedAt 通知创建时间
}

// NotificationListResponse 通知列表响应
// 用于返回用户的通知列表及总数
type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"` // Notifications 通知列表
	Total         int64                  `json:"total"`         // Total 通知总数
}

// UnreadCountResponse 未读数响应
// 用于返回用户未读通知数量统计
type UnreadCountResponse struct {
	UnreadCount int64            `json:"unreadCount"`      // UnreadCount 未读通知总数
	ByType      map[string]int64 `json:"byType,omitempty"` // ByType 按类型统计的未读数，可选字段
}

// MarkReadRequest 标记已读请求
// 用于标记通知为已读或未读
type MarkReadRequest struct {
	IsRead bool `json:"isRead" binding:"required"` // IsRead 是否已读，必填
}
