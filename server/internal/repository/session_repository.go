package repository

import (
	"context"

	"github.com/yourorg/livemix/internal/model"
	"gorm.io/gorm"
)

// ScreenShareRepository 屏幕共享仓储
type ScreenShareRepository struct {
	db *gorm.DB
}

// NewScreenShareRepository 创建屏幕共享仓储实例
func NewScreenShareRepository(db *gorm.DB) *ScreenShareRepository {
	return &ScreenShareRepository{db: db}
}

// Create 创建屏幕共享会话
func (r *ScreenShareRepository) Create(ctx context.Context, session *model.ScreenShareSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// FindByID 根据ID查询会话
func (r *ScreenShareRepository) FindByID(ctx context.Context, id uint64) (*model.ScreenShareSession, error) {
	var session model.ScreenShareSession
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// FindActiveByRoom 查询房间内正在进行的屏幕共享
func (r *ScreenShareRepository) FindActiveByRoom(ctx context.Context, roomID uint64) (*model.ScreenShareSession, error) {
	var session model.ScreenShareSession
	err := r.db.WithContext(ctx).
		Where("room_id = ? AND ended_at IS NULL", roomID).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// FindActiveByUser 查询用户正在进行的屏幕共享
func (r *ScreenShareRepository) FindActiveByUser(ctx context.Context, userID uint64) (*model.ScreenShareSession, error) {
	var session model.ScreenShareSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND ended_at IS NULL", userID).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// EndSession 结束屏幕共享会话
func (r *ScreenShareRepository) EndSession(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.ScreenShareSession{}).
		Where("id = ?", id).
		Update("ended_at", gorm.Expr("NOW()")).Error
}

// EndAllByRoom 结束房间内所有屏幕共享
func (r *ScreenShareRepository) EndAllByRoom(ctx context.Context, roomID uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.ScreenShareSession{}).
		Where("room_id = ? AND ended_at IS NULL", roomID).
		Update("ended_at", gorm.Expr("NOW()")).Error
}

// VoiceSessionRepository 语音会话仓储
type VoiceSessionRepository struct {
	db *gorm.DB
}

// NewVoiceSessionRepository 创建语音会话仓储实例
func NewVoiceSessionRepository(db *gorm.DB) *VoiceSessionRepository {
	return &VoiceSessionRepository{db: db}
}

// Create 创建语音会话
func (r *VoiceSessionRepository) Create(ctx context.Context, session *model.VoiceSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// FindActiveByRoomAndUser 查询用户在房间的活跃语音会话
func (r *VoiceSessionRepository) FindActiveByRoomAndUser(ctx context.Context, roomID, userID uint64) (*model.VoiceSession, error) {
	var session model.VoiceSession
	err := r.db.WithContext(ctx).
		Where("room_id = ? AND user_id = ? AND left_at IS NULL", roomID, userID).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// FindActiveByRoom 查询房间内所有活跃语音会话
func (r *VoiceSessionRepository) FindActiveByRoom(ctx context.Context, roomID uint64) ([]model.VoiceSession, error) {
	var sessions []model.VoiceSession
	err := r.db.WithContext(ctx).
		Where("room_id = ? AND left_at IS NULL", roomID).
		Find(&sessions).Error
	return sessions, err
}

// Leave 离开语音会话
func (r *VoiceSessionRepository) Leave(ctx context.Context, roomID, userID uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.VoiceSession{}).
		Where("room_id = ? AND user_id = ? AND left_at IS NULL", roomID, userID).
		Update("left_at", gorm.Expr("NOW()")).Error
}

// LeaveAllByUser 用户离开所有语音会话
func (r *VoiceSessionRepository) LeaveAllByUser(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.VoiceSession{}).
		Where("user_id = ? AND left_at IS NULL", userID).
		Update("left_at", gorm.Expr("NOW()")).Error
}
