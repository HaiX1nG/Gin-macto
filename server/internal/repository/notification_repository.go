package repository

import (
	"context"

	"github.com/yourorg/livemix/internal/model"
	"gorm.io/gorm"
)

// NotificationRepository 通知仓储
// 负责通知数据的持久化和查询操作
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository 创建通知仓储实例
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create 创建通知记录
// 参数：
//   - ctx: 上下文
//   - notification: 通知模型指针
//
// 返回：错误信息
func (r *NotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// FindByUserID 查询用户的通知列表
// 支持分页和类型过滤，按创建时间倒序排列
// 参数：
//   - ctx: 上下文
//   - userID: 用户ID
//   - page: 页码（从1开始）
//   - pageSize: 每页记录数
//   - notificationType: 通知类型过滤（空字符串表示不过滤）
//
// 返回：通知列表、总数、错误信息
func (r *NotificationRepository) FindByUserID(ctx context.Context, userID uint64, page, pageSize int, notificationType string) ([]model.Notification, int64, error) {
	var notifications []model.Notification
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Notification{}).Where("user_id = ?", userID)

	// 如果指定了通知类型，添加类型过滤条件
	if notificationType != "" {
		query = query.Where("type = ?", notificationType)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&notifications).Error
	return notifications, total, err
}

// FindByID 根据ID查询通知
// 参数：
//   - ctx: 上下文
//   - id: 通知ID
//
// 返回：通知模型指针、错误信息
func (r *NotificationRepository) FindByID(ctx context.Context, id uint64) (*model.Notification, error) {
	var notification model.Notification
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// MarkAsRead 标记通知为已读或未读
// 参数：
//   - ctx: 上下文
//   - id: 通知ID
//   - userID: 用户ID（用于验证通知所属权）
//   - isRead: 是否已读
//
// 返回：错误信息
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id, userID uint64, isRead bool) error {
	return r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", isRead).Error
}

// MarkAllAsRead 标记用户所有通知为已读
// 参数：
//   - ctx: 上下文
//   - userID: 用户ID
//
// 返回：错误信息
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

// GetUnreadCount 获取用户未读通知总数
// 参数：
//   - ctx: 上下文
//   - userID: 用户ID
//
// 返回：未读通知数、错误信息
func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// GetUnreadCountByType 获取用户按类型统计的未读通知数
// 参数：
//   - ctx: 上下文
//   - userID: 用户ID
//
// 返回：以通知类型为key的未读数map、错误信息
func (r *NotificationRepository) GetUnreadCountByType(ctx context.Context, userID uint64) (map[string]int64, error) {
	type countResult struct {
		Type  string `gorm:"column:type"`
		Count int64  `gorm:"column:count"`
	}

	var results []countResult
	err := r.db.WithContext(ctx).Model(&model.Notification{}).
		Select("type, COUNT(*) as count").
		Where("user_id = ? AND is_read = ?", userID, false).
		Group("type").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// 转换为 map
	countMap := make(map[string]int64, len(results))
	for _, r := range results {
		countMap[r.Type] = r.Count
	}
	return countMap, nil
}
