package repository

import (
	"context"

	"github.com/yourorg/livemix/internal/model"
	"gorm.io/gorm"
)

// PlaylistRepository 播放列表仓储
type PlaylistRepository struct {
	db *gorm.DB
}

// NewPlaylistRepository 创建播放列表仓储实例
func NewPlaylistRepository(db *gorm.DB) *PlaylistRepository {
	return &PlaylistRepository{db: db}
}

// Create 创建播放项
func (r *PlaylistRepository) Create(ctx context.Context, item *model.PlaylistItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// FindByID 根据ID查询播放项
func (r *PlaylistRepository) FindByID(ctx context.Context, id uint64) (*model.PlaylistItem, error) {
	var item model.PlaylistItem
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// FindByRoom 查询房间的播放列表
func (r *PlaylistRepository) FindByRoom(ctx context.Context, roomID uint64) ([]model.PlaylistItem, error) {
	var items []model.PlaylistItem
	err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("play_order ASC").
		Find(&items).Error
	return items, err
}

// FindWaitingByRoom 查询房间等待播放的项目
func (r *PlaylistRepository) FindWaitingByRoom(ctx context.Context, roomID uint64) ([]model.PlaylistItem, error) {
	var items []model.PlaylistItem
	err := r.db.WithContext(ctx).
		Where("room_id = ? AND status = ?", roomID, 0).
		Order("play_order ASC").
		Find(&items).Error
	return items, err
}

// FindPlayingByRoom 查询房间正在播放的项目
func (r *PlaylistRepository) FindPlayingByRoom(ctx context.Context, roomID uint64) (*model.PlaylistItem, error) {
	var item model.PlaylistItem
	err := r.db.WithContext(ctx).
		Where("room_id = ? AND status = ?", roomID, 1).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Update 更新播放项
func (r *PlaylistRepository) Update(ctx context.Context, item *model.PlaylistItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// UpdateStatus 更新播放项状态
func (r *PlaylistRepository) UpdateStatus(ctx context.Context, id uint64, status int8) error {
	return r.db.WithContext(ctx).
		Model(&model.PlaylistItem{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// Delete 删除播放项
func (r *PlaylistRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.PlaylistItem{}, id).Error
}

// GetMaxOrder 获取房间播放列表最大顺序号
func (r *PlaylistRepository) GetMaxOrder(ctx context.Context, roomID uint64) (uint32, error) {
	var maxOrder uint32
	err := r.db.WithContext(ctx).
		Model(&model.PlaylistItem{}).
		Where("room_id = ?", roomID).
		Select("COALESCE(MAX(play_order), 0)").
		Scan(&maxOrder).Error
	return maxOrder, err
}

// Reorder 按传入的 itemIDs 顺序重置播放项的 play_order
// 数组下标+1 作为新的 play_order，仅更新属于该房间的播放项
func (r *PlaylistRepository) Reorder(ctx context.Context, roomID uint64, itemIDs []uint64) error {
	if len(itemIDs) == 0 {
		return nil
	}
	for order, itemID := range itemIDs {
		if err := r.db.WithContext(ctx).
			Model(&model.PlaylistItem{}).
			Where("id = ? AND room_id = ?", itemID, roomID).
			Update("play_order", uint32(order+1)).Error; err != nil {
			return err
		}
	}
	return nil
}

// ChatMessageRepository 聊天消息仓储
type ChatMessageRepository struct {
	db *gorm.DB
}

// NewChatMessageRepository 创建聊天消息仓储实例
func NewChatMessageRepository(db *gorm.DB) *ChatMessageRepository {
	return &ChatMessageRepository{db: db}
}

// Create 创建消息
func (r *ChatMessageRepository) Create(ctx context.Context, msg *model.ChatMessage) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

// FindByRoom 查询房间消息列表
func (r *ChatMessageRepository) FindByRoom(ctx context.Context, roomID uint64, page, pageSize int) ([]model.ChatMessage, int64, error) {
	var messages []model.ChatMessage
	var total int64

	query := r.db.WithContext(ctx).Model(&model.ChatMessage{}).Where("room_id = ?", roomID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&messages).Error
	return messages, total, err
}

// FindRecentByRoom 查询房间最近消息
func (r *ChatMessageRepository) FindRecentByRoom(ctx context.Context, roomID uint64, limit int) ([]model.ChatMessage, error) {
	var messages []model.ChatMessage
	err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// FindByRoomWithFilter 带筛选条件的查询房间消息列表
func (r *ChatMessageRepository) FindByRoomWithFilter(ctx context.Context, roomID uint64, senderID uint64, messageType int8, startTime, endTime string, page, pageSize int) ([]model.ChatMessage, int64, error) {
	var messages []model.ChatMessage
	var total int64

	query := r.db.WithContext(ctx).Model(&model.ChatMessage{}).Where("room_id = ?", roomID)

	if senderID > 0 {
		query = query.Where("sender_user_id = ?", senderID)
	}
	if messageType > 0 {
		query = query.Where("message_type = ?", messageType)
	}
	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&messages).Error
	return messages, total, err
}

// FindUserHistoryMessages 查询用户参与房间的历史消息
func (r *ChatMessageRepository) FindUserHistoryMessages(ctx context.Context, userID uint64, page, pageSize int) ([]struct {
	model.ChatMessage
	RoomName string
}, int64, error) {
	var results []struct {
		model.ChatMessage
		RoomName string
	}
	var total int64

	// 查询用户参与过的房间ID
	subQuery := r.db.WithContext(ctx).
		Model(&model.RoomParticipant{}).
		Select("DISTINCT room_id").
		Where("user_id = ?", userID)

	// 统计总数
	countQuery := r.db.WithContext(ctx).
		Model(&model.ChatMessage{}).
		Where("room_id IN (?)", subQuery)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询消息
	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Table("chat_messages").
		Select("chat_messages.*, rooms.room_name").
		Joins("LEFT JOIN rooms ON chat_messages.room_id = rooms.id").
		Where("chat_messages.room_id IN (?)", subQuery).
		Order("chat_messages.created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Scan(&results).Error

	return results, total, err
}

// SearchMessages 搜索消息
func (r *ChatMessageRepository) SearchMessages(ctx context.Context, query string, roomID uint64, userID uint64, page, pageSize int) ([]model.ChatMessage, int64, error) {
	var messages []model.ChatMessage
	var total int64

	// 查询用户参与过的房间ID
	subQuery := r.db.WithContext(ctx).
		Model(&model.RoomParticipant{}).
		Select("DISTINCT room_id").
		Where("user_id = ?", userID)

	// 构建搜索查询
	searchQuery := r.db.WithContext(ctx).
		Model(&model.ChatMessage{}).
		Where("content LIKE ?", "%"+query+"%").
		Where("room_id IN (?)", subQuery)

	// 如果指定了房间ID，则只在该房间搜索
	if roomID > 0 {
		searchQuery = searchQuery.Where("room_id = ?", roomID)
	}

	// 统计总数
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询消息
	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Model(&model.ChatMessage{}).
		Where("content LIKE ?", "%"+query+"%").
		Where("room_id IN (?)", subQuery).
		Where(roomID > 0, "room_id = ?", roomID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&messages).Error

	return messages, total, err
}

// FindByID 根据消息ID查询消息
func (r *ChatMessageRepository) FindByID(ctx context.Context, messageID uint64) (*model.ChatMessage, error) {
	var msg model.ChatMessage
	err := r.db.WithContext(ctx).Where("id = ?", messageID).First(&msg).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// UpdateContent 更新消息内容（仅更新 content 列，不依赖 updated_at 列）
func (r *ChatMessageRepository) UpdateContent(ctx context.Context, messageID uint64, content string) error {
	return r.db.WithContext(ctx).
		Model(&model.ChatMessage{}).
		Where("id = ?", messageID).
		Update("content", content).Error
}

// Delete 删除消息
func (r *ChatMessageRepository) Delete(ctx context.Context, messageID uint64) error {
	return r.db.WithContext(ctx).Delete(&model.ChatMessage{}, messageID).Error
}
