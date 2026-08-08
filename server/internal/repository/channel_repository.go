package repository

import (
	"context"

	"github.com/yourorg/livemix/internal/model"
	"gorm.io/gorm"
)

// ChannelRepository 频道仓储
type ChannelRepository struct {
	db *gorm.DB
}

// NewChannelRepository 创建频道仓储实例
func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

// Create 创建频道
func (r *ChannelRepository) Create(ctx context.Context, channel *model.Channel) error {
	return r.db.WithContext(ctx).Create(channel).Error
}

// CreateWithDB 使用指定 DB 创建频道（用于事务）
func (r *ChannelRepository) CreateWithDB(db *gorm.DB, channel *model.Channel) error {
	return db.Create(channel).Error
}

// FindByID 根据ID查询频道
func (r *ChannelRepository) FindByID(ctx context.Context, id uint64) (*model.Channel, error) {
	var channel model.Channel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&channel).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// FindByServer 查询服务器的所有频道（树形列表）
func (r *ChannelRepository) FindByServer(ctx context.Context, serverID uint64) ([]model.Channel, error) {
	var channels []model.Channel
	err := r.db.WithContext(ctx).
		Where("server_id = ?", serverID).
		Order("position ASC").
		Find(&channels).Error
	return channels, err
}

// Update 更新频道
func (r *ChannelRepository) Update(ctx context.Context, channel *model.Channel) error {
	return r.db.WithContext(ctx).Save(channel).Error
}

// Delete 删除频道
func (r *ChannelRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Channel{}, id).Error
}

// Reorder 批量更新频道排序
func (r *ChannelRepository) Reorder(ctx context.Context, serverID uint64, orders []struct {
	ID       uint64
	Position int
}) error {
	if len(orders) == 0 {
		return nil
	}
	for _, order := range orders {
		if err := r.db.WithContext(ctx).
			Model(&model.Channel{}).
			Where("id = ? AND server_id = ?", order.ID, serverID).
			Update("position", order.Position).Error; err != nil {
			return err
		}
	}
	return nil
}

// DeleteByServer 删除服务器下的所有频道
func (r *ChannelRepository) DeleteByServer(ctx context.Context, serverID uint64) error {
	return r.db.WithContext(ctx).
		Where("server_id = ?", serverID).
		Delete(&model.Channel{}).Error
}
