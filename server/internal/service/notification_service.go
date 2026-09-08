package service

import (
	"context"

	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/model"
	"github.com/yourorg/livemix/internal/repository"
	"github.com/yourorg/livemix/pkg/errcode"
	"github.com/yourorg/livemix/pkg/util"
)

// NotificationService 通知服务
// 负责通知相关业务逻辑处理
type NotificationService struct {
	notificationRepo *repository.NotificationRepository
}

// NewNotificationService 创建通知服务实例
func NewNotificationService(notificationRepo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
	}
}

// GetNotifications 获取用户的通知列表
// 支持分页和类型过滤
// 参数：
//   - ctx: 上下文
//   - userID: 用户ID
//   - page: 页码（从1开始）
//   - pageSize: 每页记录数
//   - notificationType: 通知类型过滤（空字符串表示不过滤）
//
// 返回：通知列表响应、错误信息
func (s *NotificationService) GetNotifications(ctx context.Context, userID uint64, page, pageSize int, notificationType string) (*dto.NotificationListResponse, error) {
	// 参数校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	notifications, total, err := s.notificationRepo.FindByUserID(ctx, userID, page, pageSize, notificationType)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询通知失败")
	}

	// 如果没有通知，返回空列表
	if len(notifications) == 0 {
		return &dto.NotificationListResponse{
			Notifications: []dto.NotificationResponse{},
			Total:         0,
		}, nil
	}

	// 组装响应数据
	var responses []dto.NotificationResponse
	for _, n := range notifications {
		responses = append(responses, dto.NotificationResponse{
			ID:        n.ID,
			Type:      string(n.Type),
			SourceID:  n.SourceID,
			Title:     n.Title,
			Content:   n.Content,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &dto.NotificationListResponse{
		Notifications: responses,
		Total:         total,
	}, nil
}

// MarkAsRead 标记通知为已读或未读
// 验证通知所属权后更新已读状态
// 参数：
//   - ctx: 上下文
//   - userID: 当前用户ID
//   - notificationID: 通知ID
//   - isRead: 是否已读
//
// 返回：错误信息
func (s *NotificationService) MarkAsRead(ctx context.Context, userID, notificationID uint64, isRead bool) error {
	// 验证通知是否存在且属于当前用户
	notification, err := s.notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		return errcode.ErrNotificationNotFound
	}
	if notification.UserID != userID {
		return errcode.ErrForbidden.WithMessage("无权操作此通知")
	}

	// 更新已读状态
	if err = s.notificationRepo.MarkAsRead(ctx, notificationID, userID, isRead); err != nil {
		return errcode.ErrDBError.WithMessage("更新通知状态失败")
	}

	return nil
}

// MarkAllAsRead 标记用户所有通知为已读
// 参数：
//   - ctx: 上下文
//   - userID: 用户ID
//
// 返回：错误信息
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uint64) error {
	if err := s.notificationRepo.MarkAllAsRead(ctx, userID); err != nil {
		return errcode.ErrDBError.WithMessage("标记全部已读失败")
	}
	return nil
}

// GetUnreadCount 获取用户未读通知数
// 支持按类型统计
// 参数：
//   - ctx: 上下文
//   - userID: 用户ID
//
// 返回：未读数响应、错误信息
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uint64) (*dto.UnreadCountResponse, error) {
	// 获取总未读数
	total, err := s.notificationRepo.GetUnreadCount(ctx, userID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询未读数失败")
	}

	// 获取按类型统计的未读数
	byType, err := s.notificationRepo.GetUnreadCountByType(ctx, userID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询分类未读数失败")
	}

	return &dto.UnreadCountResponse{
		UnreadCount: total,
		ByType:      byType,
	}, nil
}

// CreateNotification 创建通知
// 供其他服务调用，用于生成通知记录
// 参数：
//   - ctx: 上下文
//   - userID: 接收通知的用户ID
//   - sourceID: 来源ID（消息ID/请求ID等）
//   - nType: 通知类型
//   - title: 通知标题
//   - content: 通知内容
//
// 返回：错误信息
func (s *NotificationService) CreateNotification(ctx context.Context, userID, sourceID uint64, nType model.NotificationType, title, content string) error {
	// XSS过滤：对标题和内容进行HTML转义，防止存储型XSS攻击
	title = util.TrimAndEscape(title)
	content = util.TrimAndEscape(content)

	notification := &model.Notification{
		UserID:   userID,
		Type:     nType,
		SourceID: sourceID,
		Title:    title,
		Content:  content,
		IsRead:   false,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return errcode.ErrDBError.WithMessage("创建通知失败")
	}
	return nil
}
