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

// UpdateFriendRequestWithDB 使用指定 DB 更新好友请求（用于事务）
// 注意：传入的 db 应通过 TransactionContext.DB() 获取，该 DB 已包含 context
// 使用场景：在事务中更新好友请求状态，与创建好友关系在同一事务中执行
func (r *FriendRepository) UpdateFriendRequestWithDB(db *gorm.DB, req *model.FriendRequest) error {
	return db.Save(req).Error
}

// CreateFriendshipWithDB 使用指定 DB 创建好友关系（用于事务）
// 注意：传入的 db 应通过 TransactionContext.DB() 获取，该 DB 已包含 context
// 使用场景：在事务中创建双向好友关系，与更新好友请求状态在同一事务中执行
// 该方法会创建两条好友关系记录（双向），确保操作的原子性
func (r *FriendRepository) CreateFriendshipWithDB(db *gorm.DB, userID, friendID uint64) error {
	// 创建 user -> friend
	friendship1 := &model.Friendship{
		UserID:   userID,
		FriendID: friendID,
	}
	if err := db.Create(friendship1).Error; err != nil {
		return err
	}

	// 创建 friend -> user
	friendship2 := &model.Friendship{
		UserID:   friendID,
		FriendID: userID,
	}
	if err := db.Create(friendship2).Error; err != nil {
		return err
	}

	return nil
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

// GetLastMessagesBatch 批量获取多个会话的最后一条消息（解决 N+1 查询问题）
// 参数 userPairs 为 [(userID, otherUserID), ...] 的切片
// 返回以 "userID_otherUserID" 为 key 的最后消息 map（key 中较小的 ID 在前）
func (r *FriendRepository) GetLastMessagesBatch(ctx context.Context, userID uint64, otherUserIDs []uint64) (map[uint64]*model.PrivateMessage, error) {
	if len(otherUserIDs) == 0 {
		return make(map[uint64]*model.PrivateMessage), nil
	}

	// 使用子查询获取每个会话的最后一条消息
	// 查询条件：(sender_id = userID AND receiver_id IN otherUserIDs) OR (sender_id IN otherUserIDs AND receiver_id = userID)
	var messages []model.PrivateMessage
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT pm.* FROM private_messages pm
			INNER JOIN (
				SELECT
					CASE
						WHEN sender_id = ? THEN receiver_id
						ELSE sender_id
					END as other_user_id,
					MAX(created_at) as max_created_at
				FROM private_messages
				WHERE (sender_id = ? AND receiver_id IN ?)
				   OR (sender_id IN ? AND receiver_id = ?)
				GROUP BY other_user_id
			) latest ON (
				(pm.sender_id = ? AND pm.receiver_id = latest.other_user_id AND pm.created_at = latest.max_created_at)
				OR (pm.receiver_id = ? AND pm.sender_id = latest.other_user_id AND pm.created_at = latest.max_created_at)
			)
		`, userID, userID, otherUserIDs, otherUserIDs, userID, userID, userID).
		Scan(&messages).Error
	if err != nil {
		return nil, err
	}

	// 转换为 map，key 为对方用户 ID
	msgMap := make(map[uint64]*model.PrivateMessage, len(messages))
	for i := range messages {
		var otherUserID uint64
		if messages[i].SenderID == userID {
			otherUserID = messages[i].ReceiverID
		} else {
			otherUserID = messages[i].SenderID
		}
		msgMap[otherUserID] = &messages[i]
	}
	return msgMap, nil
}

// GetUnreadCountBatch 批量获取来自多个用户的未读消息数（解决 N+1 查询问题）
// 返回以 senderID 为 key 的未读数 map
func (r *FriendRepository) GetUnreadCountBatch(ctx context.Context, receiverID uint64, senderIDs []uint64) (map[uint64]int64, error) {
	if len(senderIDs) == 0 {
		return make(map[uint64]int64), nil
	}

	type unreadCount struct {
		SenderID uint64 `gorm:"column:sender_id"`
		Count    int64  `gorm:"column:count"`
	}

	var results []unreadCount
	err := r.db.WithContext(ctx).Model(&model.PrivateMessage{}).
		Select("sender_id, COUNT(*) as count").
		Where("sender_id IN ? AND receiver_id = ? AND is_read = ?", senderIDs, receiverID, false).
		Group("sender_id").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// 转换为 map
	countMap := make(map[uint64]int64, len(results))
	for _, r := range results {
		countMap[r.SenderID] = r.Count
	}
	return countMap, nil
}
