package repository

import (
	"context"

	"github.com/yourorg/livemix/internal/model"
	"gorm.io/gorm"
)

// FriendRepository 好友仓储
type FriendRepository struct {
	db *gorm.DB
}

// NewFriendRepository 创建好友仓储实例
func NewFriendRepository(db *gorm.DB) *FriendRepository {
	return &FriendRepository{db: db}
}

// CreateFriendRequest 创建好友请求
func (r *FriendRepository) CreateFriendRequest(ctx context.Context, req *model.FriendRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

// FindFriendRequestByID 根据ID查询好友请求
func (r *FriendRepository) FindFriendRequestByID(ctx context.Context, id uint64) (*model.FriendRequest, error) {
	var req model.FriendRequest
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// FindPendingRequestByUsers 查询两个用户之间的待处理请求
func (r *FriendRepository) FindPendingRequestByUsers(ctx context.Context, senderID, receiverID uint64) (*model.FriendRequest, error) {
	var req model.FriendRequest
	err := r.db.WithContext(ctx).
		Where("sender_id = ? AND receiver_id = ? AND status = ?", senderID, receiverID, model.FriendRequestPending).
		First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// FindAnyPendingRequestByUsers 查询两个用户之间的任意待处理请求（双向）
func (r *FriendRepository) FindAnyPendingRequestByUsers(ctx context.Context, userID1, userID2 uint64) (*model.FriendRequest, error) {
	var req model.FriendRequest
	err := r.db.WithContext(ctx).
		Where("((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)) AND status = ?",
			userID1, userID2, userID2, userID1, model.FriendRequestPending).
		First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// UpdateFriendRequest 更新好友请求
func (r *FriendRepository) UpdateFriendRequest(ctx context.Context, req *model.FriendRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

// FindPendingRequestsByReceiver 查询用户收到的待处理请求
func (r *FriendRepository) FindPendingRequestsByReceiver(ctx context.Context, receiverID uint64, page, pageSize int) ([]model.FriendRequest, int64, error) {
	var requests []model.FriendRequest
	var total int64

	query := r.db.WithContext(ctx).Model(&model.FriendRequest{}).Where("receiver_id = ? AND status = ?", receiverID, model.FriendRequestPending)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&requests).Error
	return requests, total, err
}

// FindSentRequestsByUser 查询用户发送的请求
func (r *FriendRepository) FindSentRequestsByUser(ctx context.Context, senderID uint64, page, pageSize int) ([]model.FriendRequest, int64, error) {
	var requests []model.FriendRequest
	var total int64

	query := r.db.WithContext(ctx).Model(&model.FriendRequest{}).Where("sender_id = ?", senderID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&requests).Error
	return requests, total, err
}

// CreateFriendship 创建好友关系（双向）
func (r *FriendRepository) CreateFriendship(ctx context.Context, userID, friendID uint64) error {
	// 使用事务创建双向关系
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建 user -> friend
		friendship1 := &model.Friendship{
			UserID:   userID,
			FriendID: friendID,
		}
		if err := tx.Create(friendship1).Error; err != nil {
			return err
		}

		// 创建 friend -> user
		friendship2 := &model.Friendship{
			UserID:   friendID,
			FriendID: userID,
		}
		if err := tx.Create(friendship2).Error; err != nil {
			return err
		}

		return nil
	})
}

// FindFriendship 查询好友关系
func (r *FriendRepository) FindFriendship(ctx context.Context, userID, friendID uint64) (*model.Friendship, error) {
	var friendship model.Friendship
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		First(&friendship).Error
	if err != nil {
		return nil, err
	}
	return &friendship, nil
}

// DeleteFriendship 删除好友关系（双向）
func (r *FriendRepository) DeleteFriendship(ctx context.Context, userID, friendID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除 user -> friend
		if err := tx.Where("user_id = ? AND friend_id = ?", userID, friendID).Delete(&model.Friendship{}).Error; err != nil {
			return err
		}

		// 删除 friend -> user
		if err := tx.Where("user_id = ? AND friend_id = ?", friendID, userID).Delete(&model.Friendship{}).Error; err != nil {
			return err
		}

		return nil
	})
}

// FindFriendsByUserID 查询用户的好友列表
func (r *FriendRepository) FindFriendsByUserID(ctx context.Context, userID uint64, page, pageSize int) ([]model.Friendship, int64, error) {
	var friendships []model.Friendship
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Friendship{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&friendships).Error
	return friendships, total, err
}

// IsFriend 检查是否为好友
func (r *FriendRepository) IsFriend(ctx context.Context, userID, friendID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Friendship{}).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Count(&count).Error
	return count > 0, err
}

// CreatePrivateMessage 创建私聊消息
func (r *FriendRepository) CreatePrivateMessage(ctx context.Context, msg *model.PrivateMessage) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

// FindPrivateMessages 查询两个用户之间的私聊消息
func (r *FriendRepository) FindPrivateMessages(ctx context.Context, userID1, userID2 uint64, page, pageSize int) ([]model.PrivateMessage, int64, error) {
	var messages []model.PrivateMessage
	var total int64

	query := r.db.WithContext(ctx).Model(&model.PrivateMessage{}).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID1, userID2, userID2, userID1)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&messages).Error
	return messages, total, err
}

// MarkMessagesAsRead 标记消息为已读
func (r *FriendRepository) MarkMessagesAsRead(ctx context.Context, senderID, receiverID uint64) error {
	return r.db.WithContext(ctx).Model(&model.PrivateMessage{}).
		Where("sender_id = ? AND receiver_id = ? AND is_read = ?", senderID, receiverID, false).
		Update("is_read", true).Error
}

// GetUnreadCount 获取未读消息数
func (r *FriendRepository) GetUnreadCount(ctx context.Context, receiverID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PrivateMessage{}).
		Where("receiver_id = ? AND is_read = ?", receiverID, false).
		Count(&count).Error
	return count, err
}

// GetUnreadCountFromUser 获取来自特定用户的未读消息数
func (r *FriendRepository) GetUnreadCountFromUser(ctx context.Context, senderID, receiverID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PrivateMessage{}).
		Where("sender_id = ? AND receiver_id = ? AND is_read = ?", senderID, receiverID, false).
		Count(&count).Error
	return count, err
}

// GetLastMessage 获取两个用户之间的最后一条消息
func (r *FriendRepository) GetLastMessage(ctx context.Context, userID1, userID2 uint64) (*model.PrivateMessage, error) {
	var msg model.PrivateMessage
	err := r.db.WithContext(ctx).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID1, userID2, userID2, userID1).
		Order("created_at DESC").
		First(&msg).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &msg, nil
}

// FindConversations 查询用户的所有会话
func (r *FriendRepository) FindConversations(ctx context.Context, userID uint64) ([]uint64, error) {
	var userIDs []uint64

	// 查询所有发送过消息给当前用户的用户
	err := r.db.WithContext(ctx).Model(&model.PrivateMessage{}).
		Select("DISTINCT sender_id").
		Where("receiver_id = ?", userID).
		Pluck("sender_id", &userIDs).Error
	if err != nil {
		return nil, err
	}

	// 查询当前用户发送过消息的所有用户
	var receiverIDs []uint64
	err = r.db.WithContext(ctx).Model(&model.PrivateMessage{}).
		Select("DISTINCT receiver_id").
		Where("sender_id = ?", userID).
		Pluck("receiver_id", &receiverIDs).Error
	if err != nil {
		return nil, err
	}

	// 合并并去重
	userIDMap := make(map[uint64]bool)
	for _, id := range userIDs {
		userIDMap[id] = true
	}
	for _, id := range receiverIDs {
		userIDMap[id] = true
	}

	var result []uint64
	for id := range userIDMap {
		result = append(result, id)
	}

	return result, nil
}
