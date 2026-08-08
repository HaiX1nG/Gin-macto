package repository

import (
	"context"

	"github.com/yourorg/livemix/internal/model"
	"gorm.io/gorm"
)

// MessageRepository 频道消息仓储
type MessageRepository struct {
	db *gorm.DB
}

// NewMessageRepository 创建频道消息仓储实例
func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create 创建消息
func (r *MessageRepository) Create(ctx context.Context, msg *model.ChannelMessage) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

// FindByID 根据ID查询消息
func (r *MessageRepository) FindByID(ctx context.Context, id uint64) (*model.ChannelMessage, error) {
	var msg model.ChannelMessage
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&msg).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// FindByChannel 分页查询频道消息
func (r *MessageRepository) FindByChannel(ctx context.Context, channelID uint64, page, pageSize int) ([]model.ChannelMessage, int64, error) {
	var messages []model.ChannelMessage
	var total int64

	query := r.db.WithContext(ctx).Model(&model.ChannelMessage{}).Where("channel_id = ?", channelID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&messages).Error
	return messages, total, err
}

// FindRecent 查询频道最近消息
func (r *MessageRepository) FindRecent(ctx context.Context, channelID uint64, limit int) ([]model.ChannelMessage, error) {
	var messages []model.ChannelMessage
	err := r.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// Update 更新消息
func (r *MessageRepository) Update(ctx context.Context, msg *model.ChannelMessage) error {
	return r.db.WithContext(ctx).Save(msg).Error
}

// UpdateContent 更新消息内容
func (r *MessageRepository) UpdateContent(ctx context.Context, messageID uint64, content string) error {
	return r.db.WithContext(ctx).
		Model(&model.ChannelMessage{}).
		Where("id = ?", messageID).
		Update("content", content).Error
}

// Delete 删除消息
func (r *MessageRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.ChannelMessage{}, id).Error
}

// FindPinned 查询频道置顶消息
func (r *MessageRepository) FindPinned(ctx context.Context, channelID uint64) ([]model.ChannelMessage, error) {
	var messages []model.ChannelMessage
	err := r.db.WithContext(ctx).
		Where("channel_id = ? AND is_pinned = ?", channelID, true).
		Order("created_at DESC").
		Find(&messages).Error
	return messages, err
}

// SearchByChannel 在频道中搜索消息
func (r *MessageRepository) SearchByChannel(ctx context.Context, query string, channelID uint64, page, pageSize int) ([]model.ChannelMessage, int64, error) {
	var messages []model.ChannelMessage
	var total int64

	searchQuery := r.db.WithContext(ctx).
		Model(&model.ChannelMessage{}).
		Where("channel_id = ? AND content LIKE ?", channelID, "%"+query+"%")

	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := searchQuery.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&messages).Error
	return messages, total, err
}

// UpdatePinStatus 更新消息置顶状态
func (r *MessageRepository) UpdatePinStatus(ctx context.Context, messageID uint64, isPinned bool) error {
	return r.db.WithContext(ctx).
		Model(&model.ChannelMessage{}).
		Where("id = ?", messageID).
		Update("is_pinned", isPinned).Error
}

// ReactionRepository 表情反应仓储
type ReactionRepository struct {
	db *gorm.DB
}

// NewReactionRepository 创建表情反应仓储实例
func NewReactionRepository(db *gorm.DB) *ReactionRepository {
	return &ReactionRepository{db: db}
}

// Add 添加表情反应
func (r *ReactionRepository) Add(ctx context.Context, reaction *model.MessageReaction) error {
	return r.db.WithContext(ctx).Create(reaction).Error
}

// Remove 移除表情反应
func (r *ReactionRepository) Remove(ctx context.Context, messageID, userID uint64, emoji string) error {
	return r.db.WithContext(ctx).
		Where("message_id = ? AND user_id = ? AND emoji = ?", messageID, userID, emoji).
		Delete(&model.MessageReaction{}).Error
}

// ListByMessage 查询消息的所有表情反应
func (r *ReactionRepository) ListByMessage(ctx context.Context, messageID uint64) ([]model.MessageReaction, error) {
	var reactions []model.MessageReaction
	err := r.db.WithContext(ctx).
		Where("message_id = ?", messageID).
		Find(&reactions).Error
	return reactions, err
}
