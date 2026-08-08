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

// FindByChannel 查询频道的播放列表
func (r *PlaylistRepository) FindByChannel(ctx context.Context, channelID uint64) ([]model.PlaylistItem, error) {
	var items []model.PlaylistItem
	err := r.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Order("play_order ASC").
		Find(&items).Error
	return items, err
}

// FindWaitingByChannel 查询频道等待播放的项目
func (r *PlaylistRepository) FindWaitingByChannel(ctx context.Context, channelID uint64) ([]model.PlaylistItem, error) {
	var items []model.PlaylistItem
	err := r.db.WithContext(ctx).
		Where("channel_id = ? AND status = ?", channelID, 0).
		Order("play_order ASC").
		Find(&items).Error
	return items, err
}

// FindPlayingByChannel 查询频道正在播放的项目
func (r *PlaylistRepository) FindPlayingByChannel(ctx context.Context, channelID uint64) (*model.PlaylistItem, error) {
	var item model.PlaylistItem
	err := r.db.WithContext(ctx).
		Where("channel_id = ? AND status = ?", channelID, 1).
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

// GetMaxOrder 获取频道播放列表最大顺序号
func (r *PlaylistRepository) GetMaxOrder(ctx context.Context, channelID uint64) (uint32, error) {
	var maxOrder uint32
	err := r.db.WithContext(ctx).
		Model(&model.PlaylistItem{}).
		Where("channel_id = ?", channelID).
		Select("COALESCE(MAX(play_order), 0)").
		Scan(&maxOrder).Error
	return maxOrder, err
}

// Reorder 按传入的 itemIDs 顺序重置播放项的 play_order
func (r *PlaylistRepository) Reorder(ctx context.Context, channelID uint64, itemIDs []uint64) error {
	if len(itemIDs) == 0 {
		return nil
	}
	for order, itemID := range itemIDs {
		if err := r.db.WithContext(ctx).
			Model(&model.PlaylistItem{}).
			Where("id = ? AND channel_id = ?", itemID, channelID).
			Update("play_order", uint32(order+1)).Error; err != nil {
			return err
		}
	}
	return nil
}
