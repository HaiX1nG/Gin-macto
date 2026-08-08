package repository

import (
	"context"
	"time"

	"github.com/yourorg/livemix/internal/model"
	"gorm.io/gorm"
)

// VoiceRepository 语音仓储
// 管理 VoiceParticipant(实时状态)、VoiceSession(历史)、ScreenShareSession
type VoiceRepository struct {
	db *gorm.DB
}

// NewVoiceRepository 创建语音仓储实例
func NewVoiceRepository(db *gorm.DB) *VoiceRepository {
	return &VoiceRepository{db: db}
}

// ==================== VoiceParticipant 实时状态 ====================

// UpsertParticipant 创建或更新语音参与者（upsert）
func (r *VoiceRepository) UpsertParticipant(ctx context.Context, p *model.VoiceParticipant) error {
	return r.db.WithContext(ctx).
		Where("channel_id = ? AND user_id = ?", p.ChannelID, p.UserID).
		Assign(map[string]any{
			"is_muted":     p.IsMuted,
			"is_deafened":  p.IsDeafened,
			"is_speaking":  p.IsSpeaking,
			"volume":       p.Volume,
			"joined_at":    p.JoinedAt,
		}).
		FirstOrCreate(p).Error
}

// DeleteParticipant 删除语音参与者
func (r *VoiceRepository) DeleteParticipant(ctx context.Context, channelID, userID uint64) error {
	return r.db.WithContext(ctx).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Delete(&model.VoiceParticipant{}).Error
}

// FindParticipantsByChannel 查询频道的所有语音参与者
func (r *VoiceRepository) FindParticipantsByChannel(ctx context.Context, channelID uint64) ([]model.VoiceParticipant, error) {
	var participants []model.VoiceParticipant
	err := r.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Order("joined_at ASC").
		Find(&participants).Error
	return participants, err
}

// UpdateMute 更新参与者静音状态
func (r *VoiceRepository) UpdateMute(ctx context.Context, channelID, userID uint64, isMuted bool) error {
	return r.db.WithContext(ctx).
		Model(&model.VoiceParticipant{}).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Update("is_muted", isMuted).Error
}

// ==================== VoiceSession 历史记录 ====================

// CreateSession 创建语音会话历史记录
func (r *VoiceRepository) CreateSession(ctx context.Context, session *model.VoiceSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// EndSession 结束语音会话（设置 left_at）
func (r *VoiceRepository) EndSession(ctx context.Context, channelID, userID uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.VoiceSession{}).
		Where("channel_id = ? AND user_id = ? AND left_at IS NULL", channelID, userID).
		Update("left_at", time.Now()).Error
}

// ==================== ScreenShareSession ====================

// CreateScreenShare 创建屏幕共享会话
func (r *VoiceRepository) CreateScreenShare(ctx context.Context, session *model.ScreenShareSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// FindActiveScreenShareByChannel 查询频道正在进行的屏幕共享
func (r *VoiceRepository) FindActiveScreenShareByChannel(ctx context.Context, channelID uint64) (*model.ScreenShareSession, error) {
	var session model.ScreenShareSession
	err := r.db.WithContext(ctx).
		Where("channel_id = ? AND ended_at IS NULL", channelID).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// EndScreenShare 结束用户的屏幕共享
func (r *VoiceRepository) EndScreenShare(ctx context.Context, channelID, userID uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.ScreenShareSession{}).
		Where("channel_id = ? AND user_id = ? AND ended_at IS NULL", channelID, userID).
		Update("ended_at", time.Now()).Error
}
