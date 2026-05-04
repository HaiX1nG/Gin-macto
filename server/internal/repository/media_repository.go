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